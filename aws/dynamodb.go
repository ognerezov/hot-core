package aws

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ognerezov/hot-core/api"
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

// GetOnboardingConfig loads the quest configuration for the supplied sort key.
// It returns nil if the item does not exist.
func GetOnboardingConfig(ctx context.Context, client *dynamodb.Client, table, sortKey string) ([]api.Quest, error) {
	item, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(table),
		Key:       dynamoKey(api.Onboarding, sortKey),
	})
	if err != nil {
		return nil, fmt.Errorf("get onboarding configuration: %w", err)
	}
	if len(item.Item) == 0 {
		return nil, nil
	}

	var cfg api.OnboardingConfig
	if err = attributevalue.UnmarshalMap(item.Item, &cfg); err != nil {
		return nil, fmt.Errorf("decode onboarding configuration: %w", err)
	}
	return cfg.Quests, nil
}

// GetOnboardingProgress returns the user's quest statuses, or nil if absent.
func GetOnboardingProgress(ctx context.Context, client *dynamodb.Client, table, username string) (map[string]string, error) {
	item, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName:      aws.String(table),
		Key:            dynamoKey(username, api.OnboardingDone),
		ConsistentRead: aws.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("get onboarding progress: %w", err)
	}

	var onboarding api.OnboardingProgress
	if err = attributevalue.UnmarshalMap(item.Item, &onboarding); err != nil {
		return nil, fmt.Errorf("decode onboarding progress: %w", err)
	}
	return onboarding.Progress, nil
}

// CompleteQuest atomically completes a started quest. It returns nil when the
// quest is unknown or the user's progress does not allow completion.
func CompleteQuest(ctx context.Context, client *dynamodb.Client, table, username, questID string, quests []api.Quest) (*api.QuestCompleteResult, error) {
	for i, quest := range quests {
		if quest.Id != questID {
			continue
		}

		_, err := client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
			TableName:           aws.String(table),
			Key:                 dynamoKey(username, api.OnboardingDone),
			ConditionExpression: aws.String("#progress.#quest = :started"),
			UpdateExpression:    aws.String("SET #progress.#quest = :completed"),
			ExpressionAttributeNames: map[string]string{
				"#progress": "Progress",
				"#quest":    questID,
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":started":   &types.AttributeValueMemberS{Value: api.Started},
				":completed": &types.AttributeValueMemberS{Value: api.Completed},
			},
		})
		var conditionFailed *types.ConditionalCheckFailedException
		if errors.As(err, &conditionFailed) {
			return nil, nil
		}
		if err != nil {
			return nil, fmt.Errorf("complete onboarding quest: %w", err)
		}

		result := &api.QuestCompleteResult{Reward: quest.Reward}
		if i+1 < len(quests) {
			next := quests[i+1]
			result.Next = &next
		}
		return result, nil
	}
	return nil, nil
}

func dynamoKey(partitionKey, sortKey string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"PK": &types.AttributeValueMemberS{Value: partitionKey},
		"SK": &types.AttributeValueMemberS{Value: sortKey},
	}
}
