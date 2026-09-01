package handlers

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
)

func NotFoundHandler(_ context.Context, _ events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return ErrorResponse(fmt.Errorf("not found"), 404)
}
