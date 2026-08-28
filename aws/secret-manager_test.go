package aws

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/jsii-runtime-go"
	"github.com/stretchr/testify/assert"
)

func TestGetAndSetSmClient(t *testing.T) {
	oldClient := smClient
	defer func() { smClient = oldClient }()

	mock := &secretsmanager.Client{}
	SetSmClient(mock)

	client := GetSmClient()
	assert.Equal(t, mock, client)

	clientFromPrivate := getSmClient()
	assert.Equal(t, mock, clientFromPrivate)
}

func TestGetSmClientInRegion(t *testing.T) {
	oldClient := smClient
	defer func() { smClient = oldClient }()

	mock := &secretsmanager.Client{}
	SetSmClient(mock)

	client := GetSmClientInRegion("")
	assert.Equal(t, mock, client)

	regClient1 := GetSmClientInRegion("eu-west-1")
	assert.NotNil(t, regClient1)
	assert.Equal(t, "eu-west-1", regClient1.Options().Region)

	regClient2 := GetSmClientInRegion("eu-west-1")
	assert.Equal(t, regClient1, regClient2)
}

func TestLoadSecretAndRegional(t *testing.T) {
	oldClient := smClient
	defer func() { smClient = oldClient }()

	mockClient := &http.Client{
		Transport: MockTransport(func(req *http.Request) (*http.Response, error) {
			body := `{"Name":"my-secret","SecretString":"secret_val"}`
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
	SetSmClient(secretsmanager.NewFromConfig(cfg))

	val, err := LoadSecret("my-secret")
	assert.NoError(t, err)
	assert.NotNil(t, val)
	assert.Equal(t, "secret_val", *val)

	valReg, err := LoadSecretInRegion("", "my-secret")
	assert.NoError(t, err)
	assert.NotNil(t, valReg)
	assert.Equal(t, "secret_val", *valReg)
}

func TestSecretExistsAndRegional(t *testing.T) {
	oldClient := smClient
	defer func() { smClient = oldClient }()

	mockClient := &http.Client{
		Transport: MockTransport(func(req *http.Request) (*http.Response, error) {
			body := `{"Name":"existing-secret","SecretString":"secret_val"}`
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
	SetSmClient(secretsmanager.NewFromConfig(cfg))

	exists, err := SecretExists("existing-secret")
	assert.NoError(t, err)
	assert.True(t, exists)

	existsReg, err := SecretExistsInRegion("", "existing-secret")
	assert.NoError(t, err)
	assert.True(t, existsReg)
}

func TestCreateAndSaveSecretInRegion(t *testing.T) {
	mockClient := &http.Client{
		Transport: MockTransport(func(req *http.Request) (*http.Response, error) {
			body := `{"ARN":"arn:aws:secretsmanager:us-west-2:123456789012:secret:test","Name":"test"}`
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
		config.WithRegion("us-west-2"),
	)
	smClientsMu.Lock()
	smClientsMap["us-west-2"] = secretsmanager.NewFromConfig(cfg)
	smClientsMu.Unlock()

	err := CreateSecretInRegion("us-west-2", "test", jsii.String("val"))
	assert.NoError(t, err)

	err = SaveSecretInRegion("us-west-2", "test", jsii.String("new_val"))
	assert.NoError(t, err)
}

func TestAnyFromSecretInRegion(t *testing.T) {
	type Sample struct {
		Key string `json:"key"`
	}

	mockClient := &http.Client{
		Transport: MockTransport(func(req *http.Request) (*http.Response, error) {
			body := `{"Name":"json-secret","SecretString":"{\"key\":\"sample_value\"}"}`
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
	smClientsMu.Lock()
	smClientsMap["eu-central-1"] = secretsmanager.NewFromConfig(cfg)
	smClientsMu.Unlock()

	var sample Sample
	err := AnyFromSecretInRegion("eu-central-1", "json-secret", &sample)
	assert.NoError(t, err)
	assert.Equal(t, "sample_value", sample.Key)
}

func TestGetDbCredentialsInRegion(t *testing.T) {
	mockClient := &http.Client{
		Transport: MockTransport(func(req *http.Request) (*http.Response, error) {
			body := `{"Name":"db-secret","SecretString":"{\"username\":\"admin\",\"password\":\"secret_pass\"}"}`
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
	smClientsMu.Lock()
	smClientsMap["eu-central-1"] = secretsmanager.NewFromConfig(cfg)
	smClientsMu.Unlock()

	creds, err := GetDbCredentialsInRegion("eu-central-1", "db-secret")
	assert.NoError(t, err)
	assert.NotNil(t, creds)
	assert.Equal(t, "admin", creds.Username)
	assert.Equal(t, "secret_pass", creds.Password)
}

func TestSaveFileAsSecretInRegion(t *testing.T) {
	mockClient := &http.Client{
		Transport: MockTransport(func(req *http.Request) (*http.Response, error) {
			body := `{"ARN":"arn:aws:secretsmanager:eu-central-1:123456789012:secret:file-secret","Name":"file-secret"}`
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
	smClientsMu.Lock()
	smClientsMap["eu-central-1"] = secretsmanager.NewFromConfig(cfg)
	smClientsMu.Unlock()

	tmpFile, err := os.CreateTemp("", "test-secret-*.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString(`{"hello":"world"}`)
	tmpFile.Close()

	err = SaveFileAsSecretInRegion("eu-central-1", "file-secret", tmpFile.Name())
	assert.NoError(t, err)
}
