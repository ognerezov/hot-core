package aws

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/stretchr/testify/assert"
)

type MockTransport func(req *http.Request) (*http.Response, error)

func (m MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m(req)
}

func TestSQSMessage(t *testing.T) {
	// Сохраняем и восстанавливаем клиент после тестов
	oldClient := sqsClient
	defer func() { sqsClient = oldClient }()

	t.Run("Successful send", func(t *testing.T) {
		mockClient := &http.Client{
			Transport: MockTransport(func(req *http.Request) (*http.Response, error) {
				body := `{"MessageId":"msg-123","MD5OfMessageBody":"817648cf4463d7bf0d5059321a5241fa"}`
				header := make(http.Header)
				header.Set("Content-Type", "application/x-amz-json-1.0")
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
		SetSQSClient(sqs.NewFromConfig(cfg))

		msg := UserMessage[string]{
			Payload:  "hello",
			Provider: "test",
			UserID:   "user1",
		}
		output, err := SQSMessage("https://sqs.us-east-1.amazonaws.com/123/my-queue", msg)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, "msg-123", *output.MessageId)
	})

	t.Run("Send error", func(t *testing.T) {
		mockClient := &http.Client{
			Transport: MockTransport(func(req *http.Request) (*http.Response, error) {
				body := `{"__type":"com.amazonaws.sqs#InvalidParameterValue","message":"Value (zero) for parameter MessageGroupId is invalid."}`
				header := make(http.Header)
				header.Set("Content-Type", "application/x-amz-json-1.0")
				return &http.Response{
					StatusCode: 400,
					Header:     header,
					Body:       io.NopCloser(strings.NewReader(body)),
				}, nil
			}),
		}

		cfg, _ := config.LoadDefaultConfig(context.Background(),
			config.WithHTTPClient(mockClient),
			config.WithRegion("us-east-1"),
		)
		SetSQSClient(sqs.NewFromConfig(cfg))

		msg := UserMessage[string]{Payload: "hello"}
		_, err := SQSMessage("https://sqs.us-east-1.amazonaws.com/123/my-queue", msg)

		assert.Error(t, err)
	})
}

func TestQueue_Send(t *testing.T) {
	oldClient := sqsClient
	defer func() { sqsClient = oldClient }()

	mockClient := &http.Client{
		Transport: MockTransport(func(req *http.Request) (*http.Response, error) {
			// MD5 for {"payload":42,"provider":"","userId":""}
			body := `{"MessageId":"queue-msg","MD5OfMessageBody":"5f4e9a98b31ab4bb5a37768080863f06"}`
			header := make(http.Header)
			header.Set("Content-Type", "application/x-amz-json-1.0")
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
	SetSQSClient(sqs.NewFromConfig(cfg))

	q := NewQueue[int]("https://sqs.us-east-1.amazonaws.com/123/my-queue")
	msg := UserMessage[int]{Payload: 42}
	output, err := q.Send(msg)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "queue-msg", *output.MessageId)
}

func TestGetSQSClient(t *testing.T) {
	oldClient := sqsClient
	defer func() { sqsClient = oldClient }()

	mock := &sqs.Client{}
	SetSQSClient(mock)

	client := GetSQSClient()
	if client != mock {
		t.Error("GetSQSClient should return the client set by SetSQSClient")
	}
}
