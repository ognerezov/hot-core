package aws

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/ognerezov/hot-core/api"
)

func ExtractCognitoMainAndRegionalSub(req events.APIGatewayV2HTTPRequest) api.Profile {
	regionalSub := req.RequestContext.Authorizer.JWT.Claims["sub"]
	userID := req.RequestContext.Authorizer.JWT.Claims["main_sub"]
	if userID == "" {
		userID = regionalSub
	}

	return api.Profile{
		RegionalSub:   regionalSub,
		MainSub:       userID,
		Name:          req.RequestContext.Authorizer.JWT.Claims["name"],
		Email:         req.RequestContext.Authorizer.JWT.Claims["email"],
		EmailVerified: req.RequestContext.Authorizer.JWT.Claims["email_verified"] == "true",
	}
}
