package aws

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"
	"time"

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
	return cfg.Items, nil
}

// GetOnboardingProgress decodes the user's onboarding progress.
func GetOnboardingProgress(ctx context.Context, client *dynamodb.Client, table, username string) (*api.OnboardingProgress, error) {
	item, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName:      aws.String(table),
		Key:            dynamoKey(username, api.OnboardingDone),
		ConsistentRead: aws.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("get onboarding progress: %w", err)
	}

	var onboarding api.OnboardingProgress
	if err := attributevalue.UnmarshalMap(item.Item, &onboarding); err != nil {
		return nil, fmt.Errorf("decode onboarding progress: %w", err)
	}
	return &onboarding, nil
}

// StartOnboarding creates the user's progress and queues the first quest's reward.
// It returns nil if onboarding already exists or there are no quests.
func StartOnboarding(ctx context.Context, client *dynamodb.Client, table, username string, quests []api.Quest) (*api.QuestCompleteResult, error) {
	if len(quests) == 0 {
		return nil, nil
	}

	first := quests[0]
	if first.Id == "" {
		return nil, errors.New("first onboarding quest has no ID")
	}
	progress := map[string]string{first.Id: api.ProcessingReward}
	result := &api.QuestCompleteResult{Reward: first.Reward, Next: []api.Quest{}}
	if len(quests) > 1 {
		next := quests[1]
		if next.Id == "" || next.Id == first.Id {
			return nil, errors.New("second onboarding quest must have a distinct ID")
		}
		progress[next.Id] = api.Started
		result.Next = append(result.Next, next)
	}

	item, err := attributevalue.MarshalMap(api.OnboardingProgress{
		PK:       username,
		SK:       api.OnboardingDone,
		Progress: progress,
		PendingRewards: []api.ReceivedRewards{{
			Reward:    first.Reward,
			Date:      time.Now().UTC().Format(time.RFC3339Nano),
			QuestId:   first.Id,
			QuestName: first.Name,
		}},
		ReceivedRewards: []api.ReceivedRewards{},
	})
	if err != nil {
		return nil, fmt.Errorf("encode initial onboarding progress: %w", err)
	}
	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:                aws.String(table),
		Item:                     item,
		ConditionExpression:      aws.String("attribute_not_exists(#pk)"),
		ExpressionAttributeNames: map[string]string{"#pk": "PK"},
	})
	var conditionFailed *types.ConditionalCheckFailedException
	if errors.As(err, &conditionFailed) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("start onboarding: %w", err)
	}
	return result, nil
}

// CompleteQuest moves a started quest to reward processing. It returns nil
// when the quest is unknown or a reward cannot be started.
func CompleteQuest(ctx context.Context, client *dynamodb.Client, table, username, questID string, quests []api.Quest) (*api.QuestCompleteResult, error) {
	for i, quest := range quests {
		if quest.Id != questID {
			continue
		}

		rewards, err := attributevalue.Marshal([]api.ReceivedRewards{{
			Reward:    quest.Reward,
			Date:      time.Now().UTC().Format(time.RFC3339Nano),
			QuestId:   questID,
			QuestName: quest.Name,
		}})
		if err != nil {
			return nil, fmt.Errorf("encode pending onboarding reward: %w", err)
		}
		_, err = client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
			TableName:           aws.String(table),
			Key:                 dynamoKey(username, api.OnboardingDone),
			ConditionExpression: aws.String("#progress.#quest = :started"),
			UpdateExpression:    aws.String("SET #progress.#quest = :processing, #pending = list_append(if_not_exists(#pending, :emptyList), :rewards)"),
			ExpressionAttributeNames: map[string]string{
				"#progress": "Progress",
				"#quest":    questID,
				"#pending":  "PendingRewards",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":started":    &types.AttributeValueMemberS{Value: api.Started},
				":processing": &types.AttributeValueMemberS{Value: api.ProcessingReward},
				":emptyList":  &types.AttributeValueMemberL{},
				":rewards":    rewards,
			},
		})
		var conditionFailed *types.ConditionalCheckFailedException
		if errors.As(err, &conditionFailed) {
			return nil, nil
		}
		if err != nil {
			return nil, fmt.Errorf("complete onboarding quest: %w", err)
		}

		result := &api.QuestCompleteResult{Reward: quest.Reward, Next: []api.Quest{}}
		if i+1 < len(quests) {
			result.Next = append(result.Next, quests[i+1])
		}
		return result, nil
	}
	return nil, nil
}

// OnRewardProcessed moves the supplied pending rewards into history.
// It returns false if the batch is empty or is no longer pending.
func OnRewardProcessed(ctx context.Context, client *dynamodb.Client, table string, onboarding api.OnboardingProgress) (bool, error) {
	batch := onboarding.PendingRewards
	if len(batch) == 0 {
		return false, nil
	}
	if onboarding.PK == "" || onboarding.SK != api.OnboardingDone {
		return false, errors.New("invalid onboarding progress key")
	}

	pending, err := attributevalue.Marshal(batch)
	if err != nil {
		return false, fmt.Errorf("encode pending onboarding rewards: %w", err)
	}
	received := slices.Clone(batch)
	processedAt := time.Now().UTC().Format(time.RFC3339Nano)
	for i := range received {
		received[i].Date = processedAt
	}
	rewards, err := attributevalue.Marshal(received)
	if err != nil {
		return false, fmt.Errorf("encode received onboarding rewards: %w", err)
	}
	progress, err := attributevalue.Marshal(onboarding.Progress)
	if err != nil {
		return false, fmt.Errorf("encode onboarding progress: %w", err)
	}

	key := dynamoKey(onboarding.PK, onboarding.SK)
	historyExists := onboarding.ReceivedRewards != nil
	var remaining []types.AttributeValue
	for {
		completed := make(map[string]types.AttributeValue)
		if existing, ok := progress.(*types.AttributeValueMemberM); ok {
			maps.Copy(completed, existing.Value)
		}
		for _, reward := range batch {
			if reward.QuestId != "" {
				completed[reward.QuestId] = &types.AttributeValueMemberS{Value: api.Completed}
			}
		}

		condition := "#pending = :pending"
		values := map[string]types.AttributeValue{
			":pending":   pending,
			":completed": &types.AttributeValueMemberM{Value: completed},
			":rewards":   rewards,
		}
		switch progress.(type) {
		case nil, *types.AttributeValueMemberNULL:
			condition += " AND (attribute_not_exists(#progress) OR attribute_type(#progress, :nullType))"
			values[":nullType"] = &types.AttributeValueMemberS{Value: "NULL"}
		default:
			condition += " AND #progress = :progress"
			values[":progress"] = progress
		}

		update := "SET #progress = :completed, #received = "
		if historyExists {
			condition += " AND attribute_type(#received, :listType)"
			values[":listType"] = &types.AttributeValueMemberS{Value: "L"}
			update += "list_append(#received, :rewards)"
		} else {
			condition += " AND (attribute_not_exists(#received) OR attribute_type(#received, :nullType))"
			values[":nullType"] = &types.AttributeValueMemberS{Value: "NULL"}
			update += ":rewards"
		}
		if len(remaining) == 0 {
			update += " REMOVE #pending"
		} else {
			update += ", #pending = :remaining"
			values[":remaining"] = &types.AttributeValueMemberL{Value: remaining}
		}

		_, err = client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
			TableName:           aws.String(table),
			Key:                 key,
			ConditionExpression: aws.String(condition),
			UpdateExpression:    aws.String(update),
			ExpressionAttributeNames: map[string]string{
				"#progress": "Progress",
				"#pending":  "PendingRewards",
				"#received": "ReceivedRewards",
			},
			ExpressionAttributeValues:           values,
			ReturnValuesOnConditionCheckFailure: types.ReturnValuesOnConditionCheckFailureAllOld,
		})
		if err == nil {
			return true, nil
		}
		var conflict *types.ConditionalCheckFailedException
		if !errors.As(err, &conflict) {
			return false, fmt.Errorf("record processed onboarding rewards: %w", err)
		}

		current, ok := conflict.Item["PendingRewards"].(*types.AttributeValueMemberL)
		if !ok || len(current.Value) < len(batch) {
			return false, nil
		}
		var currentBatch []api.ReceivedRewards
		if err := attributevalue.UnmarshalList(current.Value[:len(batch)], &currentBatch); err != nil {
			return false, fmt.Errorf("decode pending onboarding rewards after conflict: %w", err)
		}
		if !slices.Equal(batch, currentBatch) {
			return false, nil
		}

		// Retry the same batch against the returned item, preserving newly queued rewards.
		pending = current
		remaining = current.Value[len(batch):]
		progress = conflict.Item["Progress"]
		switch conflict.Item["ReceivedRewards"].(type) {
		case nil, *types.AttributeValueMemberNULL:
			historyExists = false
		case *types.AttributeValueMemberL:
			historyExists = true
		default:
			return false, errors.New("invalid ReceivedRewards attribute: expected a list")
		}
	}
}

func dynamoKey(partitionKey, sortKey string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"PK": &types.AttributeValueMemberS{Value: partitionKey},
		"SK": &types.AttributeValueMemberS{Value: sortKey},
	}
}
