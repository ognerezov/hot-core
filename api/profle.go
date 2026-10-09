package api

type Profile struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	RegionalSub   string `json:"regional_sub"`
	MainSub       string `json:"main_sub"`
	Name          string `json:"name"`
}
