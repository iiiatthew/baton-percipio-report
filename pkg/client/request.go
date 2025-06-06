package client

import (
	"context"
	"fmt"
	"net/http"
	liburl "net/url"
	"strconv"
	"time"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

func (c *Client) getUrl(
	path string,
	queryParameters map[string]any,
) *liburl.URL {
	params := liburl.Values{}
	for key, valueAny := range queryParameters {
		switch value := valueAny.(type) {
		case string:
			params.Add(key, value)
		case int:
			params.Add(key, strconv.Itoa(value))
		case bool:
			params.Add(key, strconv.FormatBool(value))
		default:
			continue
		}
	}

	output := c.baseUrl.JoinPath(fmt.Sprintf(path, c.organizationId))
	output.RawQuery = params.Encode()
	return output
}

// WithBearerToken - TODO(marcos): move this function to `baton-sdk`.
func WithBearerToken(token string) uhttp.RequestOption {
	return uhttp.WithHeader("Authorization", fmt.Sprintf("Bearer %s", token))
}

func (c *Client) get(
	ctx context.Context,
	path string,
	queryParameters map[string]any,
	target any,
) (
	*http.Response,
	*v2.RateLimitDescription,
	error,
) {
	return c.doRequest(
		ctx,
		http.MethodGet,
		path,
		queryParameters,
		nil,
		&target,
	)
}

func (c *Client) post(
	ctx context.Context,
	path string,
	body interface{},
	target interface{},
) (
	*http.Response,
	*v2.RateLimitDescription,
	error,
) {
	return c.doRequest(
		ctx,
		http.MethodPost,
		path,
		nil,
		body,
		&target,
	)
}

func (c *Client) doRequest(
	ctx context.Context,
	method string,
	path string,
	queryParameters map[string]any,
	payload any,
	target any,
) (
	*http.Response,
	*v2.RateLimitDescription,
	error,
) {
	logger := ctxzap.Extract(ctx)
	startTime := time.Now()
	
	options := []uhttp.RequestOption{
		uhttp.WithAcceptJSONHeader(),
		WithBearerToken(c.bearerToken),
	}
	if payload != nil {
		options = append(options, uhttp.WithJSONBody(payload))
	}

	url := c.getUrl(path, queryParameters)
	
	logger.Debug("Making API request",
		zap.String("method", method),
		zap.String("endpoint", url.String()),
		zap.String("path", path),
		zap.Any("query_params", queryParameters),
		zap.Bool("has_payload", payload != nil))

	request, err := c.wrapper.NewRequest(ctx, method, url, options...)
	if err != nil {
		logger.Error("Failed to create request",
			zap.Error(err),
			zap.String("method", method),
			zap.String("endpoint", url.String()))
		return nil, nil, err
	}

	var ratelimitData v2.RateLimitDescription
	response, err := c.wrapper.Do(
		request,
		uhttp.WithRatelimitData(&ratelimitData),
		uhttp.WithJSONResponse(target),
	)
	
	if err != nil {
		logger.Error("API request failed",
			zap.Error(err),
			zap.String("method", method),
			zap.String("endpoint", url.String()),
			zap.Duration("duration", time.Since(startTime)))
		return response, &ratelimitData, fmt.Errorf("error making %s request to %s: %w", method, url, err)
	}

	logger.Debug("API request completed",
		zap.String("method", method),
		zap.String("endpoint", url.String()),
		zap.Int("status_code", response.StatusCode),
		zap.Duration("duration", time.Since(startTime)),
		zap.Int("rate_limit_remaining", int(ratelimitData.Remaining)),
		zap.Int64("rate_limit_reset_at", ratelimitData.ResetAt.Seconds))

	return response, &ratelimitData, nil
}
