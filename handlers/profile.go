package handlers

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ognerezov/hot-core/aws"
)

func GetProfileHandler(_ context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	res := aws.ExtractCognitoMainAndRegionalSub(req)

	return JsonResponse(res, 200)
}
