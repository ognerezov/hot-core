package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/rs/zerolog/log"
)

var (
	sqsClient *sqs.Client
	sqsOnce   sync.Once
)

type UserMessage[T any] struct {
	Payload    T      `json:"payload"`
	Provider   string `json:"provider"`
	UserID     string `json:"userId"`
	RawPayload string `json:"rawPayload"`
}

func SetSQSClient(client *sqs.Client) {
	sqsClient = client
}

func GetSQSClient() *sqs.Client {
	sqsOnce.Do(func() {
		if sqsClient != nil {
			return
		}

		cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(os.Getenv("CDK_REGION")))
		if err != nil {
			panic(fmt.Errorf("failed to load AWS config for SQS: %v", err))
		}

		sqsClient = sqs.NewFromConfig(cfg)
	})

	return sqsClient
}

func SQSMessage(url string, msg any) (*sqs.SendMessageOutput, error) {
	sqsBody, err := json.Marshal(msg)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal SQS payload")
		return nil, err
	}
	client := GetSQSClient()

	log.Info().Str("queueUrl", url).Msgf("Sending message to SQS %s", string(sqsBody))
	return client.SendMessage(context.Background(), &sqs.SendMessageInput{
		QueueUrl:    aws.String(url),
		MessageBody: aws.String(string(sqsBody)),
	})
}

type Queue[T any] struct {
	url string
}

func NewQueue[T any](url string) *Queue[T] {
	return &Queue[T]{url: url}
}

func (q *Queue[T]) SendUserMessage(msg UserMessage[T]) (*sqs.SendMessageOutput, error) {
	return SQSMessage(q.url, msg)
}

func (q *Queue[T]) Send(msg T) (*sqs.SendMessageOutput, error) {
	return SQSMessage(q.url, msg)
}
