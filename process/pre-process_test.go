package process

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/ognerezov/hot-core/aws"
	"github.com/stretchr/testify/assert"
)

type mockTransport func(req *http.Request) (*http.Response, error)

func (m mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m(req)
}

func createMockSsmClient(region string, paramResponses map[string]string) *ssm.Client {
	mockClient := &http.Client{
		Transport: mockTransport(func(req *http.Request) (*http.Response, error) {
			buf := new(strings.Builder)
			_, _ = io.Copy(buf, req.Body)
			bodyStr := buf.String()

			for paramName, val := range paramResponses {
				if strings.Contains(bodyStr, fmt.Sprintf(`"Name":"%s"`, paramName)) ||
					strings.Contains(bodyStr, fmt.Sprintf(`"Name": "%s"`, paramName)) {
					respBody := fmt.Sprintf(`{"Parameter":{"Name":"%s","Value":"%s","Type":"SecureString"}}`, paramName, val)
					header := make(http.Header)
					header.Set("Content-Type", "application/x-amz-json-1.1")
					return &http.Response{
						StatusCode: 200,
						Header:     header,
						Body:       io.NopCloser(strings.NewReader(respBody)),
					}, nil
				}
			}

			header := make(http.Header)
			header.Set("Content-Type", "application/x-amz-json-1.1")
			return &http.Response{
				StatusCode: 400,
				Header:     header,
				Body:       io.NopCloser(strings.NewReader(`{"__type":"ParameterNotFound","message":"Parameter not found."}`)),
			}, nil
		}),
	}

	cfg, _ := config.LoadDefaultConfig(context.Background(),
		config.WithHTTPClient(mockClient),
		config.WithRegion(region),
	)
	return ssm.NewFromConfig(cfg)
}

func TestPreProcess(t *testing.T) {
	mockClient := createMockSsmClient("us-east-1", map[string]string{
		GoogleClientId:     "google-id-123",
		GoogleClientSecret: "google-secret-456",
	})
	aws.SetSsmClient(mockClient)

	params, err := PreProcess([]string{GoogleClientId, GoogleClientSecret})
	assert.NoError(t, err)
	assert.NotNil(t, params)
	assert.Equal(t, "google-id-123", *params[GoogleClientId])
	assert.Equal(t, "google-secret-456", *params[GoogleClientSecret])
}

func TestPreProcessInRegion(t *testing.T) {
	mockClient := createMockSsmClient("eu-west-1", map[string]string{
		AppleClientId: "apple-client-eu",
		AppleKeyId:    "apple-key-eu",
	})
	aws.SetSsmClientInRegion("eu-west-1", mockClient)

	params, err := PreProcessInRegion("eu-west-1", []string{AppleClientId, AppleKeyId})
	assert.NoError(t, err)
	assert.NotNil(t, params)
	assert.Equal(t, "apple-client-eu", *params[AppleClientId])
	assert.Equal(t, "apple-key-eu", *params[AppleKeyId])
}

func TestPreProcessInRegions(t *testing.T) {
	mockUS := createMockSsmClient("us-east-1", map[string]string{
		GoogleClientId: "us-google-id",
		AppleClientId:  "us-apple-id",
	})
	mockEU := createMockSsmClient("eu-central-1", map[string]string{
		GoogleClientId: "eu-google-id",
		AppleClientId:  "eu-apple-id",
	})

	aws.SetSsmClientInRegion("us-east-1", mockUS)
	aws.SetSsmClientInRegion("eu-central-1", mockEU)

	regions := []string{"us-east-1", "eu-central-1"}
	paramNames := []string{GoogleClientId, AppleClientId}

	res, err := PreProcessInRegions(regions, paramNames)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res, 2)

	assert.Equal(t, "us-google-id", *res["us-east-1"][GoogleClientId])
	assert.Equal(t, "us-apple-id", *res["us-east-1"][AppleClientId])

	assert.Equal(t, "eu-google-id", *res["eu-central-1"][GoogleClientId])
	assert.Equal(t, "eu-apple-id", *res["eu-central-1"][AppleClientId])

	// Test alias PreProcessRegions
	resAlias, err := PreProcessRegions(regions, paramNames)
	assert.NoError(t, err)
	assert.Equal(t, res, resAlias)
}

func TestPreProcessInRegions_Error(t *testing.T) {
	mockUS := createMockSsmClient("us-east-1", map[string]string{
		GoogleClientId: "us-google-id",
	})
	aws.SetSsmClientInRegion("us-east-1", mockUS)

	regions := []string{"us-east-1"}
	paramNames := []string{"non-existing-param"}

	res, err := PreProcessInRegions(regions, paramNames)
	assert.Error(t, err)
	assert.Nil(t, res)
}
