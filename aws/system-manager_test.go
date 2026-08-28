package aws

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/stretchr/testify/assert"
)

func TestGetAndSetSsmClient(t *testing.T) {
	oldClient := ssmClient
	defer func() { ssmClient = oldClient }()

	mock := &ssm.Client{}
	SetSsmClient(mock)

	client := GetSsmClient()
	assert.Equal(t, mock, client)

	clientFromPrivate := getSsmClient()
	assert.Equal(t, mock, clientFromPrivate)
}

func TestGetSsmClientInRegion(t *testing.T) {
	oldClient := ssmClient
	defer func() { ssmClient = oldClient }()

	mock := &ssm.Client{}
	SetSsmClient(mock)

	// empty region should return default client
	client := GetSsmClientInRegion("")
	assert.Equal(t, mock, client)

	// specific region returns a configured client and caches it
	regClient1 := GetSsmClientInRegion("eu-west-1")
	assert.NotNil(t, regClient1)
	assert.Equal(t, "eu-west-1", regClient1.Options().Region)

	regClient2 := GetSsmClientInRegion("eu-west-1")
	assert.Equal(t, regClient1, regClient2)
}

func TestGetSecureParameter(t *testing.T) {
	oldClient := ssmClient
	defer func() { ssmClient = oldClient }()

	mockClient := &http.Client{
		Transport: MockTransport(func(req *http.Request) (*http.Response, error) {
			body := `{"Parameter":{"Name":"/my/test/param","Value":"secret_value_123","Type":"SecureString"}}`
			header := make(http.Header)
			header.Set("Content-Type", "application/x-amz-json-1.1")
			return &http.Response{
				StatusCode: 200,
				Header:     header,
				Body:       io.NopCloser(strings.NewReader(body)),
			}, nil
		}),
	}

	cfg, _ := config.LoadDefaultConfig(context.Background(),
		config.WithHTTPClient(mockClient),
		config.WithRegion("us-east-1"),
	)
	SetSsmClient(ssm.NewFromConfig(cfg))

	val, err := GetSecureParameter("/my/test/param")
	assert.NoError(t, err)
	assert.NotNil(t, val)
	assert.Equal(t, "secret_value_123", *val)
}

func TestGetSecureParameterInRegion(t *testing.T) {
	oldClient := ssmClient
	defer func() { ssmClient = oldClient }()

	mockClient := &http.Client{
		Transport: MockTransport(func(req *http.Request) (*http.Response, error) {
			body := `{"Parameter":{"Name":"/my/test/param","Value":"regional_secret_value","Type":"SecureString"}}`
			header := make(http.Header)
			header.Set("Content-Type", "application/x-amz-json-1.1")
			return &http.Response{
				StatusCode: 200,
				Header:     header,
				Body:       io.NopCloser(strings.NewReader(body)),
			}, nil
		}),
	}

	cfg, _ := config.LoadDefaultConfig(context.Background(),
		config.WithHTTPClient(mockClient),
		config.WithRegion("eu-central-1"),
	)
	ssmClientsMu.Lock()
	ssmClientsMap["eu-central-1"] = ssm.NewFromConfig(cfg)
	ssmClientsMu.Unlock()

	val, err := GetSecureParameterInRegion("eu-central-1", "/my/test/param")
	assert.NoError(t, err)
	assert.NotNil(t, val)
	assert.Equal(t, "regional_secret_value", *val)
}

func TestPutParameter(t *testing.T) {
	oldClient := ssmClient
	defer func() { ssmClient = oldClient }()

	mockClient := &http.Client{
		Transport: MockTransport(func(req *http.Request) (*http.Response, error) {
			body := `{"Version":1}`
			header := make(http.Header)
			header.Set("Content-Type", "application/x-amz-json-1.1")
			return &http.Response{
				StatusCode: 200,
				Header:     header,
				Body:       io.NopCloser(strings.NewReader(body)),
			}, nil
		}),
	}

	cfg, _ := config.LoadDefaultConfig(context.Background(),
		config.WithHTTPClient(mockClient),
		config.WithRegion("us-east-1"),
	)
	SetSsmClient(ssm.NewFromConfig(cfg))

	err := PutParameter("/my/test/param", "my_value")
	assert.NoError(t, err)
}

func TestPutParameterInRegion(t *testing.T) {
	mockClient := &http.Client{
		Transport: MockTransport(func(req *http.Request) (*http.Response, error) {
			body := `{"Version":2}`
			header := make(http.Header)
			header.Set("Content-Type", "application/x-amz-json-1.1")
			return &http.Response{
				StatusCode: 200,
				Header:     header,
				Body:       io.NopCloser(strings.NewReader(body)),
			}, nil
		}),
	}

	cfg, _ := config.LoadDefaultConfig(context.Background(),
		config.WithHTTPClient(mockClient),
		config.WithRegion("ap-southeast-1"),
	)
	ssmClientsMu.Lock()
	ssmClientsMap["ap-southeast-1"] = ssm.NewFromConfig(cfg)
	ssmClientsMu.Unlock()

	err := PutParameterInRegion("ap-southeast-1", "/my/test/param", "my_regional_value")
	assert.NoError(t, err)
}
