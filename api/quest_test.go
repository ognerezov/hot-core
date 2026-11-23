package api

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPendingRewardSummary(t *testing.T) {
	progress := OnboardingProgress{
		PendingRewards:  []ReceivedRewards{{QuestName: "First", Reward: 25}, {QuestName: "Second", Reward: 15}},
		ReceivedRewards: []ReceivedRewards{{QuestName: "Completed", Reward: 100}},
	}
	total, names := progress.PendingRewardSummary()
	require.Equal(t, 40, total)
	require.Equal(t, []string{"First", "Second"}, names)

	total, names = (OnboardingProgress{}).PendingRewardSummary()
	require.Zero(t, total)
	require.Empty(t, names)
}
