package handlers

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/zerolog/log"
)

type LambdaHandler func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error)

type MethodRouter struct {
	Get    LambdaHandler
	Post   LambdaHandler
	Put    LambdaHandler
	Delete LambdaHandler
	Patch  LambdaHandler
}

func (m MethodRouter) Handle(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	var handler LambdaHandler
	switch req.RequestContext.HTTP.Method {
	case "GET":
		handler = m.Get
	case "POST":
		handler = m.Post
	case "PUT":
		handler = m.Put
	case "DELETE":
		handler = m.Delete
	case "PATCH":
		handler = m.Patch
	}

	if handler == nil {
		return ErrorResponse(fmt.Errorf("method not allowed"), 405)
	}

	return handler(ctx, req)
}

type Router struct {
	Handlers             map[string]LambdaHandler
	Domains              []string
	AuthenticatedHandler LambdaHandler
	InfoHandler          LambdaHandler
	notFoundHandler      LambdaHandler
}

func NewRouter(domains []string, authHandler LambdaHandler) *Router {
	return &Router{
		Handlers:             make(map[string]LambdaHandler),
		Domains:              domains,
		AuthenticatedHandler: authHandler,
		notFoundHandler:      NotFoundHandler,
	}
}

func (r *Router) SetNotFoundHandler(handler LambdaHandler) {
	r.notFoundHandler = handler
}

func (r *Router) GetNotFoundHandler() LambdaHandler {
	if r.notFoundHandler != nil {
		return r.notFoundHandler
	}
	return NotFoundHandler
}

func (r *Router) RegisterHandler(regex string, handler LambdaHandler) {
	r.Handlers[regex] = handler
}

func (r *Router) RegisterResource(regex string, router MethodRouter) {
	r.RegisterHandler(regex, router.Handle)
}

func (r *Router) MatchingHandler(_path string) *LambdaHandler {
	parts := strings.Split(_path, "?")
	path := parts[0]
	log.Info().Str("clean path", fmt.Sprintf("%v", path))
	for regex, handler := range r.Handlers {
		matcher := regexp.MustCompile(regex)
		if matcher.MatchString(path) {
			return &handler
		}
	}
	return nil
}

func (r *Router) GetHandler(ctx context.Context, req events.APIGatewayV2HTTPRequest) LambdaHandler {
	log.Info().Str("path", req.RawPath).Str("method", req.RequestContext.HTTP.Method).Msg("Searching for handler")
	if req.RequestContext.HTTP.Method == "OPTIONS" {
		return r.OptionsHandler
	}
	if !strings.HasPrefix(req.RawPath, "/pub/") {
		if r.AuthenticatedHandler != nil {
			return r.AuthenticatedHandler
		}

		log.Error().Msg("No Authentication handler found found for private method ")
		return func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
			return ErrorResponse(fmt.Errorf("unauthorized"), 401)
		}
	}
	log.Info().Str("path", req.RawPath).Msg("This is unauthenticated route")
	var handler LambdaHandler
	h := r.MatchingHandler(req.RawPath)
	if h == nil {
		log.Error().Msg("No handler for path and no info handler")
		handler = r.GetNotFoundHandler()
	} else {
		log.Info().Str("path", req.RawPath).Msg("Found handler")
		handler = *h
	}
	return handler
}

func (r *Router) WithCorsHeaders(res events.APIGatewayV2HTTPResponse, err error) (events.APIGatewayV2HTTPResponse, error) {
	headers := res.Headers
	if headers == nil {
		headers = map[string]string{}
	}
	var domains = strings.Join(r.Domains, ",")
	headers["Access-Control-Allow-Headers"] = "Authorization, Origin, Content-Type, X-Auth-Token, Set-Cookie"
	headers["Access-Control-Allow-Methods"] = "POST, GET, OPTIONS, PUT, DELETE, PATCH, HEAD"
	headers["Access-Control-Allow-Origin"] = domains
	headers["Access-Control-Allow-Credentials"] = "true"
	res.Headers = headers
	return res, err
}

func (r *Router) OptionsHandler(_ context.Context, _ events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Headers:    map[string]string{},
	}, nil
}

func (r *Router) ApiHandler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (interface{}, error) {
	log.Info().Str("path", req.RawPath).Msg(fmt.Sprintf("%v", req.PathParameters))
	handler := r.GetHandler(ctx, req)
	return r.WithCorsHeaders(handler(ctx, req))
}

func (r *Router) Start() {
	lambda.Start(r.ApiHandler)
}
