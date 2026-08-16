package aws

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
)

func TestParseS3Uri(t *testing.T) {
	tests := []struct {
		name       string
		uri        string
		wantBucket string
		wantKey    string
		wantErr    bool
	}{
		{
			name:       "Valid URI",
			uri:        "s3://my-bucket/my/key/file.jpg",
			wantBucket: "my-bucket",
			wantKey:    "my/key/file.jpg",
			wantErr:    false,
		},
		{
			name:       "Valid URI with root file",
			uri:        "s3://my-bucket/file.jpg",
			wantBucket: "my-bucket",
			wantKey:    "file.jpg",
			wantErr:    false,
		},
		{
			name:       "Invalid Prefix",
			uri:        "http://my-bucket/file.jpg",
			wantBucket: "",
			wantKey:    "",
			wantErr:    true,
		},
		{
			name:       "Missing Key",
			uri:        "s3://my-bucket",
			wantBucket: "",
			wantKey:    "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bucket, key, err := ParseS3Uri(tt.uri)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseS3Uri() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if bucket != tt.wantBucket {
				t.Errorf("ParseS3Uri() bucket = %v, want %v", bucket, tt.wantBucket)
			}
			if key != tt.wantKey {
				t.Errorf("ParseS3Uri() key = %v, want %v", key, tt.wantKey)
			}
		})
	}
}

func TestUploadToS3(t *testing.T) {
	oldS3 := s3Client
	oldTM := tmClient
	defer func() {
		s3Client = oldS3
		tmClient = oldTM
	}()

	mockClient := &http.Client{
		Transport: MockTransport(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil
		}),
	}

	cfg, _ := config.LoadDefaultConfig(context.Background(),
		config.WithHTTPClient(mockClient),
		config.WithRegion("us-east-1"),
		config.WithRequestChecksumCalculation(aws.RequestChecksumCalculationWhenRequired),
	)
	SetS3Client(s3.NewFromConfig(cfg))

	tmpFile, err := os.CreateTemp("", "test-upload")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("test content")
	tmpFile.Close()

	err = UploadToS3("my-bucket", "my-key", tmpFile.Name(), nil)
	assert.NoError(t, err)
}

func TestDownloadFromS3(t *testing.T) {
	oldS3 := s3Client
	oldTM := tmClient
	defer func() {
		s3Client = oldS3
		tmClient = oldTM
	}()

	mockClient := &http.Client{
		Transport: MockTransport(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("test content")),
			}, nil
		}),
	}

	cfg, _ := config.LoadDefaultConfig(context.Background(),
		config.WithHTTPClient(mockClient),
		config.WithRegion("us-east-1"),
		config.WithRequestChecksumCalculation(aws.RequestChecksumCalculationWhenRequired),
	)
	SetS3Client(s3.NewFromConfig(cfg))

	tmpFile, err := os.CreateTemp("", "test-download")
	assert.NoError(t, err)
	destPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(destPath)

	err = DownloadFromS3("my-bucket", "my-key", destPath)
	assert.NoError(t, err)

	content, _ := os.ReadFile(destPath)
	assert.Equal(t, "test content", string(content))
}

func TestGetObjectAsBase64(t *testing.T) {
	oldS3 := s3Client
	oldTM := tmClient
	defer func() {
		s3Client = oldS3
		tmClient = oldTM
	}()

	mockClient := &http.Client{
		Transport: MockTransport(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("hello world")),
			}, nil
		}),
	}

	cfg, _ := config.LoadDefaultConfig(context.Background(),
		config.WithHTTPClient(mockClient),
		config.WithRegion("us-east-1"),
		config.WithRequestChecksumCalculation(aws.RequestChecksumCalculationWhenRequired),
	)
	SetS3Client(s3.NewFromConfig(cfg))

	base64Str, err := GetObjectAsBase64(context.Background(), "my-bucket", "my-key")
	assert.NoError(t, err)
	assert.Equal(t, "aGVsbG8gd29ybGQ=", base64Str)
}
