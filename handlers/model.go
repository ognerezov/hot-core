package handlers

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/rs/zerolog/log"
)

func ToString(data any) string {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "serialization error"
	}
	return string(bytes)
}

func StringToJson(msg string, field string) string {
	return ToString(map[string]string{
		field: msg,
	})
}

func ErrorResponse(err error, code int) (events.APIGatewayV2HTTPResponse, error) {
	log.Info().Msg(err.Error())
	return events.APIGatewayV2HTTPResponse{
		StatusCode: code,
		Body:       StringToJson(err.Error(), "error"),
	}, err
}

func JsonResponse(data any, code int) (events.APIGatewayV2HTTPResponse, error) {
	return events.APIGatewayV2HTTPResponse{
		StatusCode: code,
		Body:       ToString(data),
	}, nil
}
