package aws

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var (
	dynamoClient *dynamodb.Client
)

// SetDynamoClient sets a custom DynamoDB client. Useful for testing.
func SetDynamoClient(client *dynamodb.Client) {
	dynamoClient = client
}

// GetDynamoClient returns a singleton DynamoDB client, initializing it if necessary.
// It uses the CDK_REGION environment variable for configuration.
func GetDynamoClient() *dynamodb.Client {
	if dynamoClient != nil {
		return dynamoClient
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(os.Getenv("CDK_REGION")))
	if err != nil {
		panic(fmt.Errorf("failed to load AWS config: %v", err))
	}

	dynamoClient = dynamodb.NewFromConfig(cfg)
	return dynamoClient
}
