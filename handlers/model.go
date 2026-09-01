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
	return WithJsonContentType(events.APIGatewayV2HTTPResponse{
		StatusCode: code,
		Body:       StringToJson(err.Error(), "error"),
	}, err)
}

func JsonResponse(data any, code int) (events.APIGatewayV2HTTPResponse, error) {
	return WithJsonContentType(events.APIGatewayV2HTTPResponse{
		StatusCode: code,
		Body:       ToString(data),
	}, nil)
}

func WithJsonContentType(res events.APIGatewayV2HTTPResponse, err error) (events.APIGatewayV2HTTPResponse, error) {
	headers := res.Headers
	if headers == nil {
		headers = map[string]string{}
	}
	headers["Content-Type"] = "application/json; charset=utf-8"
	res.Headers = headers
	return res, err
}
