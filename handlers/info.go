package handlers

import (
	"context"
	"os"
	"strings"
	"unicode"

	"github.com/aws/aws-lambda-go/events"
)

var version string

type AppInfo struct {
	Version               string            `json:"version"`
	Region                string            `json:"region"`
	CognitoDomain         string            `json:"cognito_domain"`
	CognitoClientId       string            `json:"cognito_client_id"`
	MobileCognitoClientId string            `json:"mobile_cognito_client_id"`
	Scope                 string            `json:"scope"`
	AppSyncUrl            string            `json:"app_sync_url"`
	AppSyncAuthType       string            `json:"app_sync_auth_type"`
	UserPoolId            string            `json:"user_pool_id"`
	IdentityPoolId        string            `json:"identity_pool_id"`
	Apis                  map[string]string `json:"apis,omitzero"`
}

func InfoHandler(_ context.Context, _ events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	if version == "" {
		version = os.Getenv("APP_VERSION")
		if version == "" {
			version = "unknown"
		}
	}
	info := AppInfo{
		Version:               version,
		Region:                os.Getenv("AWS_REGION"),
		CognitoDomain:         os.Getenv("COGNITO_DOMAIN"),
		CognitoClientId:       os.Getenv("COGNITO_CLIENT_ID"),
		MobileCognitoClientId: os.Getenv("MOBILE_COGNITO_CLIENT_ID"),
		Scope:                 "openid profile email",
		AppSyncUrl:            os.Getenv("APPSYNC_URL"),
		AppSyncAuthType:       "AMAZON_COGNITO_USER_POOLS",
		UserPoolId:            os.Getenv("USER_POOL_ID"),
		IdentityPoolId:        os.Getenv("IDENTITY_POOL_ID"),
		Apis:                  make(map[string]string),
	}

	for _, env := range os.Environ() {
		if key, value, found := strings.Cut(env, "="); found && strings.HasPrefix(key, "INFO_") {
			rawKey := strings.TrimPrefix(key, "INFO_")
			camelKey := toCamelCase(rawKey)
			info.Apis[camelKey] = value
		}
	}

	return JsonResponse(info, 200)
}

func toCamelCase(s string) string {
	parts := strings.Split(strings.ToLower(s), "_")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			r := []rune(parts[i])
			r[0] = unicode.ToUpper(r[0])
			parts[i] = string(r)
		}
	}
	return strings.Join(parts, "")
}
