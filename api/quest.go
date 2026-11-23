package api

const (
	Onboarding       = "ONBOARDING"
	OnboardingDone   = "ONBOARDING_DONE"
	Started          = "Started"
	Completed        = "Completed"
	ProcessingReward = "ProcessingReward"
)

type Quest struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Data        string `json:"data"`
	Reward      int    `json:"reward"`
}

type OnboardingConfig struct {
	PK     string
	SK     string
	Quests []Quest
}

type OnboardingProgress struct {
	PK              string
	SK              string
	Progress        map[string]string
	PendingRewards  []ReceivedRewards `dynamodbav:",omitempty"`
	ReceivedRewards []ReceivedRewards
}

// PendingRewardSummary returns the total and quest names in pending order.
func (p OnboardingProgress) PendingRewardSummary() (int, []string) {
	total := 0
	names := make([]string, 0, len(p.PendingRewards))
	for _, reward := range p.PendingRewards {
		total += reward.Reward
		names = append(names, reward.QuestName)
	}
	return total, names
}

type QuestCompleteResult struct {
	Reward int     `json:"reward"`
	Next   []Quest `json:"next"`
}

type ReceivedRewards struct {
	Reward    int    `json:"reward"`
	Date      string `json:"date"`
	QuestId   string `json:"questId"`
	QuestName string `json:"questName"`
}
