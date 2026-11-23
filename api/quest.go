package api

const (
	Onboarding     = "ONBOARDING"
	OnboardingDone = "ONBOARDING_DONE"
	Started        = "Started"
	Completed      = "Completed"
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
	PK       string
	SK       string
	Progress map[string]string
}

type QuestCompleteResult struct {
	Reward int    `json:"reward"`
	Next   *Quest `json:"next"`
}
