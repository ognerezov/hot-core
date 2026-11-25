package aws

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/ognerezov/hot-core/api"
	"github.com/stretchr/testify/require"
)

const conditionalFailure = `{"__type":"com.amazonaws.dynamodb.v20120810#ConditionalCheckFailedException"}`

func TestStartOnboarding(t *testing.T) {
	first := api.Quest{Id: "first", Name: "First quest", Reward: 25}
	second := api.Quest{Id: "second", Name: "Second quest", Reward: 40}
	for _, tc := range []struct {
		name     string
		quests   []api.Quest
		progress map[string]any
		next     []api.Quest
	}{
		{
			name:     "single quest",
			quests:   []api.Quest{first},
			progress: map[string]any{"first": map[string]any{"S": api.ProcessingReward}},
			next:     []api.Quest{},
		},
		{
			name:   "starts the second quest",
			quests: []api.Quest{first, second, {Id: "third"}},
			progress: map[string]any{
				"first":  map[string]any{"S": api.ProcessingReward},
				"second": map[string]any{"S": api.Started},
			},
			next: []api.Quest{second},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			client := questTestClient(t, func(operation string, body map[string]any) (int, string) {
				require.Equal(t, "DynamoDB_20120810.PutItem", operation)
				require.Equal(t, "quests", body["TableName"])
				require.Equal(t, "attribute_not_exists(#pk)", body["ConditionExpression"])
				require.Equal(t, map[string]any{"#pk": "PK"}, body["ExpressionAttributeNames"])
				item := body["Item"].(map[string]any)
				require.Equal(t, map[string]any{"M": tc.progress}, item["Progress"])
				reward := requireReward(t, item["PendingRewards"])
				require.Equal(t, first.Id, reward.QuestId)
				require.Equal(t, first.Name, reward.QuestName)
				require.Equal(t, first.Reward, reward.Reward)
				require.Equal(t, map[string]any{"L": []any{}}, item["ReceivedRewards"])
				calls++
				if calls == 2 {
					return http.StatusBadRequest, conditionalFailure
				}
				return http.StatusOK, "{}"
			})
			result, err := StartOnboarding(t.Context(), client, "quests", "alice", tc.quests)
			require.NoError(t, err)
			require.Equal(t, &api.QuestCompleteResult{Reward: first.Reward, Next: tc.next}, result)

			result, err = StartOnboarding(t.Context(), client, "quests", "alice", tc.quests)
			require.NoError(t, err)
			require.Nil(t, result)
			require.Equal(t, 2, calls)
		})
	}
}

func TestStartOnboardingSkipsEmptyQuestList(t *testing.T) {
	result, err := StartOnboarding(t.Context(), nil, "quests", "alice", nil)
	require.NoError(t, err)
	require.Nil(t, result)
}

func TestCompleteQuestStartsReward(t *testing.T) {
	calls := 0
	client := questTestClient(t, func(target string, body map[string]any) (int, string) {
		require.Equal(t, "DynamoDB_20120810.UpdateItem", target)
		require.Equal(t, "#progress.#quest = :started", body["ConditionExpression"])
		require.Equal(t, "SET #progress.#quest = :processing, #pending = list_append(if_not_exists(#pending, :emptyList), :rewards)", body["UpdateExpression"])
		names := body["ExpressionAttributeNames"].(map[string]any)
		require.Equal(t, "PendingRewards", names["#pending"])
		require.Equal(t, "first.quest", names["#quest"])
		values := body["ExpressionAttributeValues"].(map[string]any)
		require.Equal(t, map[string]any{"S": api.ProcessingReward}, values[":processing"])
		require.Equal(t, map[string]any{"S": api.Started}, values[":started"])
		reward := requireReward(t, values[":rewards"])
		require.Equal(t, "first.quest", reward.QuestId)
		require.Equal(t, "First quest", reward.QuestName)
		require.Zero(t, reward.Reward)

		calls++
		if calls == 2 {
			return http.StatusBadRequest, conditionalFailure
		}
		return http.StatusOK, "{}"
	})

	updated, err := CompleteQuest(t.Context(), client, "quests", "alice", "first.quest", "First quest")
	require.NoError(t, err)
	require.True(t, updated)

	updated, err = CompleteQuest(t.Context(), client, "quests", "alice", "first.quest", "First quest")
	require.NoError(t, err)
	require.False(t, updated)
	require.Equal(t, 2, calls)
}

func TestOnRewardProcessedCompletesBatchRegardlessOfQuestStatus(t *testing.T) {
	const queuedAt = "2026-01-01T00:00:00Z"
	onboarding := api.OnboardingProgress{
		PK: "alice",
		SK: api.OnboardingDone,
		Progress: map[string]string{
			"first.quest": api.Started,
			"unrelated":   api.ProcessingReward,
		},
		PendingRewards: []api.ReceivedRewards{
			{QuestId: "first.quest", QuestName: "First quest", Reward: 10, Date: queuedAt},
			{QuestId: "second", QuestName: "Second quest", Reward: 25, Date: queuedAt},
		},
		ReceivedRewards: []api.ReceivedRewards{},
	}

	calls := 0
	client := questTestClient(t, func(operation string, body map[string]any) (int, string) {
		calls++
		require.Equal(t, "DynamoDB_20120810.UpdateItem", operation)
		require.Equal(t, "ALL_OLD", body["ReturnValuesOnConditionCheckFailure"])
		require.Equal(t, "#pending = :pending AND #progress = :progress AND attribute_type(#received, :listType)", body["ConditionExpression"])
		require.Equal(t, "SET #progress = :completed, #received = list_append(#received, :rewards) REMOVE #pending", body["UpdateExpression"])
		values := body["ExpressionAttributeValues"].(map[string]any)
		require.Equal(t, onboarding.PendingRewards, decodeTestRewards(t, values[":pending"]))
		require.Equal(t, map[string]any{"M": map[string]any{
			"first.quest": map[string]any{"S": api.Completed},
			"second":      map[string]any{"S": api.Completed},
			"unrelated":   map[string]any{"S": api.ProcessingReward},
		}}, values[":completed"])
		rewards := decodeTestRewards(t, values[":rewards"])
		require.Len(t, rewards, 2)
		processedAt := rewards[0].Date
		require.NotEqual(t, queuedAt, processedAt)
		require.Equal(t, []api.ReceivedRewards{
			{QuestId: "first.quest", QuestName: "First quest", Reward: 10, Date: processedAt},
			{QuestId: "second", QuestName: "Second quest", Reward: 25, Date: processedAt},
		}, rewards)
		if calls == 2 {
			return http.StatusBadRequest, conditionalFailure
		}
		return http.StatusOK, "{}"
	})

	processed, err := OnRewardProcessed(t.Context(), client, "quests", onboarding)
	require.NoError(t, err)
	require.True(t, processed)
	require.Equal(t, queuedAt, onboarding.PendingRewards[0].Date)
	require.Equal(t, api.Started, onboarding.Progress["first.quest"])

	processed, err = OnRewardProcessed(t.Context(), client, "quests", onboarding)
	require.NoError(t, err)
	require.False(t, processed)
	require.Equal(t, 2, calls)
}

func TestOnRewardProcessedInitializesMissingOrNullAttributes(t *testing.T) {
	onboarding := api.OnboardingProgress{
		PK:             "alice",
		SK:             api.OnboardingDone,
		PendingRewards: []api.ReceivedRewards{{QuestId: "first", Reward: 25}},
	}
	client := questTestClient(t, func(operation string, body map[string]any) (int, string) {
		require.Equal(t, "DynamoDB_20120810.UpdateItem", operation)
		require.Equal(t, "#pending = :pending AND (attribute_not_exists(#progress) OR attribute_type(#progress, :nullType)) AND (attribute_not_exists(#received) OR attribute_type(#received, :nullType))", body["ConditionExpression"])
		require.Equal(t, "SET #progress = :completed, #received = :rewards REMOVE #pending", body["UpdateExpression"])
		values := body["ExpressionAttributeValues"].(map[string]any)
		require.Equal(t, map[string]any{"M": map[string]any{"first": map[string]any{"S": api.Completed}}}, values[":completed"])
		reward := requireReward(t, values[":rewards"])
		require.Equal(t, "first", reward.QuestId)
		require.Equal(t, 25, reward.Reward)
		return http.StatusOK, "{}"
	})
	processed, err := OnRewardProcessed(t.Context(), client, "quests", onboarding)
	require.NoError(t, err)
	require.True(t, processed)
}

func TestOnRewardProcessedPreservesConcurrentUpdates(t *testing.T) {
	onboarding := api.OnboardingProgress{
		PK:             "alice",
		SK:             api.OnboardingDone,
		PendingRewards: []api.ReceivedRewards{{QuestId: "first", Reward: 25}},
	}
	const pending = `{"L":[{"M":{"QuestId":{"S":"first"},"Reward":{"N":"25"}}},{"M":{"QuestId":{"S":"later"},"Reward":{"N":"40"}}}]}`
	const concurrentProgress = `{"M":{"later":{"S":"ProcessingReward"},"other":{"S":"Started"}}}`
	const conflictingItem = `{"__type":"com.amazonaws.dynamodb.v20120810#ConditionalCheckFailedException","Item":{"PendingRewards":` + pending + `,"Progress":` + concurrentProgress + `,"ReceivedRewards":{"L":[{"M":{"QuestId":{"S":"earlier"}}}]}}}`
	calls := 0
	client := questTestClient(t, func(operation string, body map[string]any) (int, string) {
		require.Equal(t, "DynamoDB_20120810.UpdateItem", operation)
		calls++
		if calls == 1 {
			return http.StatusBadRequest, conflictingItem
		}

		require.Equal(t, "SET #progress = :completed, #received = list_append(#received, :rewards), #pending = :remaining", body["UpdateExpression"])
		values := body["ExpressionAttributeValues"].(map[string]any)
		actualPending, err := json.Marshal(values[":pending"])
		require.NoError(t, err)
		require.JSONEq(t, pending, string(actualPending))
		actualProgress, err := json.Marshal(values[":progress"])
		require.NoError(t, err)
		require.JSONEq(t, concurrentProgress, string(actualProgress))
		require.Equal(t, map[string]any{"M": map[string]any{
			"first": map[string]any{"S": api.Completed},
			"later": map[string]any{"S": api.ProcessingReward},
			"other": map[string]any{"S": api.Started},
		}}, values[":completed"])
		actualRemaining, err := json.Marshal(values[":remaining"])
		require.NoError(t, err)
		require.JSONEq(t, `{"L":[{"M":{"QuestId":{"S":"later"},"Reward":{"N":"40"}}}]}`, string(actualRemaining))
		require.Equal(t, "first", requireReward(t, values[":rewards"]).QuestId)
		return http.StatusOK, "{}"
	})

	processed, err := OnRewardProcessed(t.Context(), client, "quests", onboarding)
	require.NoError(t, err)
	require.True(t, processed)
	require.Equal(t, 2, calls)
}

func TestOnRewardProcessedSkipsEmptyBatch(t *testing.T) {
	processed, err := OnRewardProcessed(t.Context(), nil, "quests", api.OnboardingProgress{})
	require.NoError(t, err)
	require.False(t, processed)
}

func requireReward(t *testing.T, value any) api.ReceivedRewards {
	t.Helper()
	rewards := decodeTestRewards(t, value)
	require.Len(t, rewards, 1)
	return rewards[0]
}

func decodeTestRewards(t *testing.T, value any) []api.ReceivedRewards {
	t.Helper()
	data, err := json.Marshal(value)
	require.NoError(t, err)
	attribute, err := attributevalue.UnmarshalJSON(data)
	require.NoError(t, err)
	var rewards []api.ReceivedRewards
	require.NoError(t, attributevalue.Unmarshal(attribute, &rewards))
	for _, reward := range rewards {
		_, err := time.Parse(time.RFC3339Nano, reward.Date)
		require.NoError(t, err)
	}
	return rewards
}

func questTestClient(t *testing.T, respond func(string, map[string]any) (int, string)) *dynamodb.Client {
	return dynamodb.New(dynamodb.Options{
		Region:      "us-east-1",
		Credentials: sdkaws.AnonymousCredentials{},
		HTTPClient: &http.Client{Transport: MockTransport(func(req *http.Request) (*http.Response, error) {
			var body map[string]any
			require.NoError(t, json.NewDecoder(req.Body).Decode(&body))
			expectedKey := map[string]any{
				"PK": map[string]any{"S": "alice"}, "SK": map[string]any{"S": api.OnboardingDone},
			}
			operation := req.Header.Get("X-Amz-Target")
			if operation == "DynamoDB_20120810.PutItem" {
				item := body["Item"].(map[string]any)
				require.Equal(t, expectedKey["PK"], item["PK"])
				require.Equal(t, expectedKey["SK"], item["SK"])
			} else {
				require.Equal(t, expectedKey, body["Key"])
			}
			status, payload := respond(operation, body)
			return &http.Response{
				StatusCode: status,
				Header:     http.Header{"Content-Type": {"application/x-amz-json-1.0"}},
				Body:       io.NopCloser(strings.NewReader(payload)),
			}, nil
		})},
	})
}
