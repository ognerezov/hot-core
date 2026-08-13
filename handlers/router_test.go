package handlers

import (
	"context"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestMethodRouter_Handle(t *testing.T) {
	mockHandler := func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		return events.APIGatewayV2HTTPResponse{StatusCode: 200, Body: req.RequestContext.HTTP.Method}, nil
	}

	router := MethodRouter{
		Get:    mockHandler,
		Post:   mockHandler,
		Put:    mockHandler,
		Delete: mockHandler,
		Patch:  mockHandler,
	}

	tests := []struct {
		method     string
		wantStatus int
		wantBody   string
	}{
		{"GET", 200, "GET"},
		{"POST", 200, "POST"},
		{"PUT", 200, "PUT"},
		{"DELETE", 200, "DELETE"},
		{"PATCH", 200, "PATCH"},
		{"HEAD", 405, ""}, // Not defined in MethodRouter
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			req := events.APIGatewayV2HTTPRequest{
				RequestContext: events.APIGatewayV2HTTPRequestContext{
					HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
						Method: tt.method,
					},
				},
			}
			resp, err := router.Handle(context.Background(), req)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
			if tt.wantStatus == 200 {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantBody, resp.Body)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestRouter_MatchingHandler(t *testing.T) {
	r := NewRouter([]string{"*"}, nil)
	r.RegisterHandler("^/pub/test$", func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		return events.APIGatewayV2HTTPResponse{StatusCode: 200}, nil
	})
	r.RegisterHandler("^/pub/user/\\d+$", func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		return events.APIGatewayV2HTTPResponse{StatusCode: 201}, nil
	})

	tests := []struct {
		path       string
		wantFound  bool
		wantStatus int
	}{
		{"/pub/test", true, 200},
		{"/pub/test?query=1", true, 200},
		{"/pub/user/123", true, 201},
		{"/pub/user/abc", false, 0},
		{"/other", false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			h := r.MatchingHandler(tt.path)
			if tt.wantFound {
				assert.NotNil(t, h)
				resp, _ := (*h)(context.Background(), events.APIGatewayV2HTTPRequest{})
				assert.Equal(t, tt.wantStatus, resp.StatusCode)
			} else {
				assert.Nil(t, h)
			}
		})
	}
}

func TestRouter_GetHandler(t *testing.T) {
	authHandler := func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		return events.APIGatewayV2HTTPResponse{StatusCode: 200, Body: "auth"}, nil
	}
	pubHandler := func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		return events.APIGatewayV2HTTPResponse{StatusCode: 200, Body: "pub"}, nil
	}
	infoHandler := func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		return events.APIGatewayV2HTTPResponse{StatusCode: 200, Body: "info"}, nil
	}

	r := NewRouter([]string{"*"}, authHandler)
	r.RegisterHandler("^/pub/test$", pubHandler)
	r.InfoHandler = infoHandler

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantBody   string
	}{
		{"OPTIONS request", "OPTIONS", "/any", 200, ""},
		{"Private route with auth", "GET", "/private", 200, "auth"},
		{"Public route with handler", "GET", "/pub/test", 200, "pub"},
		{"Public route fallback to info", "GET", "/pub/not-found", 200, "info"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := events.APIGatewayV2HTTPRequest{
				RawPath: tt.path,
				RequestContext: events.APIGatewayV2HTTPRequestContext{
					HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
						Method: tt.method,
					},
				},
			}
			h := r.GetHandler(context.Background(), req)
			assert.NotNil(t, h)
			resp, _ := h(context.Background(), req)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
			if tt.wantBody != "" {
				assert.Equal(t, tt.wantBody, resp.Body)
			}
		})
	}

	t.Run("Private route without auth handler", func(t *testing.T) {
		rNoAuth := NewRouter([]string{"*"}, nil)
		req := events.APIGatewayV2HTTPRequest{RawPath: "/private"}
		h := rNoAuth.GetHandler(context.Background(), req)
		resp, err := h(context.Background(), req)
		assert.Equal(t, 401, resp.StatusCode)
		assert.Error(t, err)
	})

	t.Run("Public route without handler and without info", func(t *testing.T) {
		rNoInfo := NewRouter([]string{"*"}, nil)
		req := events.APIGatewayV2HTTPRequest{RawPath: "/pub/not-found"}
		h := rNoInfo.GetHandler(context.Background(), req)
		resp, err := h(context.Background(), req)
		assert.Equal(t, 404, resp.StatusCode)
		assert.Error(t, err)
	})
}

func TestRouter_WithCorsHeaders(t *testing.T) {
	domains := []string{"http://example.com", "http://test.com"}
	r := NewRouter(domains, nil)

	res := events.APIGatewayV2HTTPResponse{StatusCode: 200}
	res, err := r.WithCorsHeaders(res, nil)

	assert.NoError(t, err)
	assert.Equal(t, "http://example.com,http://test.com", res.Headers["Access-Control-Allow-Origin"])
	assert.Equal(t, "true", res.Headers["Access-Control-Allow-Credentials"])
	assert.Contains(t, res.Headers["Access-Control-Allow-Methods"], "GET")
	assert.Contains(t, res.Headers["Access-Control-Allow-Headers"], "Authorization")
}

func TestRouter_ApiHandler(t *testing.T) {
	pubHandler := func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		return events.APIGatewayV2HTTPResponse{StatusCode: 200, Body: "ok"}, nil
	}

	r := NewRouter([]string{"*"}, nil)
	r.RegisterHandler("^/pub/ok$", pubHandler)

	req := events.APIGatewayV2HTTPRequest{
		RawPath: "/pub/ok",
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method: "GET",
			},
		},
	}

	resInterface, err := r.ApiHandler(context.Background(), req)
	assert.NoError(t, err)

	res, ok := resInterface.(events.APIGatewayV2HTTPResponse)
	assert.True(t, ok)
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "ok", res.Body)
	assert.Equal(t, "*", res.Headers["Access-Control-Allow-Origin"])
}
