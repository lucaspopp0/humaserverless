package humaserverless

import (
	"context"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHttpMiddlewareChain(t *testing.T) {
	t.Run("empty-middleware-chain", func(t *testing.T) {
		chain := HttpMiddlewareChain()

		handler := func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
			return events.APIGatewayV2HTTPResponse{
				StatusCode: 200,
				Body:       "success",
			}, nil
		}

		wrapped := chain(handler)
		ctx := context.Background()
		req := events.APIGatewayV2HTTPRequest{
			RouteKey: "GET /test",
		}

		resp, err := wrapped(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, "success", resp.Body)
	})

	t.Run("single-middleware", func(t *testing.T) {
		middlewareExecuted := false
		middleware := func(next HttpHandlerFunc) HttpHandlerFunc {
			return func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
				middlewareExecuted = true
				return next(ctx, req)
			}
		}

		chain := HttpMiddlewareChain(middleware)

		handler := func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
			return events.APIGatewayV2HTTPResponse{
				StatusCode: 200,
				Body:       "handler executed",
			}, nil
		}

		wrapped := chain(handler)
		ctx := context.Background()
		req := events.APIGatewayV2HTTPRequest{
			RouteKey: "GET /test",
		}

		resp, err := wrapped(ctx, req)

		require.NoError(t, err)
		assert.True(t, middlewareExecuted, "middleware should have been executed")
		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, "handler executed", resp.Body)
	})

	t.Run("multiple-middlewares-execution-order", func(t *testing.T) {
		executionOrder := []string{}

		middleware1 := func(next HttpHandlerFunc) HttpHandlerFunc {
			return func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
				executionOrder = append(executionOrder, "middleware1-before")
				resp, err := next(ctx, req)
				executionOrder = append(executionOrder, "middleware1-after")
				return resp, err
			}
		}

		middleware2 := func(next HttpHandlerFunc) HttpHandlerFunc {
			return func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
				executionOrder = append(executionOrder, "middleware2-before")
				resp, err := next(ctx, req)
				executionOrder = append(executionOrder, "middleware2-after")
				return resp, err
			}
		}

		middleware3 := func(next HttpHandlerFunc) HttpHandlerFunc {
			return func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
				executionOrder = append(executionOrder, "middleware3-before")
				resp, err := next(ctx, req)
				executionOrder = append(executionOrder, "middleware3-after")
				return resp, err
			}
		}

		chain := HttpMiddlewareChain(middleware1, middleware2, middleware3)

		handler := func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
			executionOrder = append(executionOrder, "handler")
			return events.APIGatewayV2HTTPResponse{
				StatusCode: 200,
				Body:       "success",
			}, nil
		}

		wrapped := chain(handler)
		ctx := context.Background()
		req := events.APIGatewayV2HTTPRequest{
			RouteKey: "GET /test",
		}

		resp, err := wrapped(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		// Middleware1 should execute first (outermost), then middleware2, then middleware3, then handler
		// Then they unwind in reverse order
		expectedOrder := []string{
			"middleware1-before",
			"middleware2-before",
			"middleware3-before",
			"handler",
			"middleware3-after",
			"middleware2-after",
			"middleware1-after",
		}
		assert.Equal(t, expectedOrder, executionOrder)
	})

	t.Run("middleware-modifies-request", func(t *testing.T) {
		middleware := func(next HttpHandlerFunc) HttpHandlerFunc {
			return func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
				// Modify the request
				if req.Headers == nil {
					req.Headers = make(map[string]string)
				}
				req.Headers["X-Custom-Header"] = "modified-by-middleware"
				return next(ctx, req)
			}
		}

		chain := HttpMiddlewareChain(middleware)

		handler := func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
			customHeader := req.Headers["X-Custom-Header"]
			return events.APIGatewayV2HTTPResponse{
				StatusCode: 200,
				Body:       customHeader,
			}, nil
		}

		wrapped := chain(handler)
		ctx := context.Background()
		req := events.APIGatewayV2HTTPRequest{
			RouteKey: "GET /test",
			Headers:  make(map[string]string),
		}

		resp, err := wrapped(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, "modified-by-middleware", resp.Body)
	})

	t.Run("middleware-modifies-response", func(t *testing.T) {
		middleware := func(next HttpHandlerFunc) HttpHandlerFunc {
			return func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
				resp, err := next(ctx, req)
				if err != nil {
					return resp, err
				}
				// Modify the response
				if resp.Headers == nil {
					resp.Headers = make(map[string]string)
				}
				resp.Headers["X-Response-Header"] = "added-by-middleware"
				resp.StatusCode = 201
				return resp, nil
			}
		}

		chain := HttpMiddlewareChain(middleware)

		handler := func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
			return events.APIGatewayV2HTTPResponse{
				StatusCode: 200,
				Body:       "original response",
			}, nil
		}

		wrapped := chain(handler)
		ctx := context.Background()
		req := events.APIGatewayV2HTTPRequest{
			RouteKey: "GET /test",
		}

		resp, err := wrapped(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode)
		assert.Equal(t, "original response", resp.Body)
		assert.Equal(t, "added-by-middleware", resp.Headers["X-Response-Header"])
	})

	t.Run("middleware-adds-headers", func(t *testing.T) {
		middleware := func(next HttpHandlerFunc) HttpHandlerFunc {
			return func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
				resp, err := next(ctx, req)
				if err != nil {
					return resp, err
				}
				if resp.Headers == nil {
					resp.Headers = make(map[string]string)
				}
				resp.Headers["X-Powered-By"] = "humaserverless"
				resp.Headers["X-Request-ID"] = "test-123"
				return resp, nil
			}
		}

		chain := HttpMiddlewareChain(middleware)

		handler := func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
			return events.APIGatewayV2HTTPResponse{
				StatusCode: 200,
				Body:       "success",
			}, nil
		}

		wrapped := chain(handler)
		ctx := context.Background()
		req := events.APIGatewayV2HTTPRequest{
			RouteKey: "GET /test",
		}

		resp, err := wrapped(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, "humaserverless", resp.Headers["X-Powered-By"])
		assert.Equal(t, "test-123", resp.Headers["X-Request-ID"])
	})

	t.Run("middleware-short-circuits", func(t *testing.T) {
		handlerExecuted := false
		middleware := func(next HttpHandlerFunc) HttpHandlerFunc {
			return func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
				// Short-circuit: return early without calling next
				return events.APIGatewayV2HTTPResponse{
					StatusCode: 403,
					Body:       "forbidden",
				}, nil
			}
		}

		chain := HttpMiddlewareChain(middleware)

		handler := func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
			handlerExecuted = true
			return events.APIGatewayV2HTTPResponse{
				StatusCode: 200,
				Body:       "should not execute",
			}, nil
		}

		wrapped := chain(handler)
		ctx := context.Background()
		req := events.APIGatewayV2HTTPRequest{
			RouteKey: "GET /test",
		}

		resp, err := wrapped(ctx, req)

		require.NoError(t, err)
		assert.False(t, handlerExecuted, "handler should not have been executed")
		assert.Equal(t, 403, resp.StatusCode)
		assert.Equal(t, "forbidden", resp.Body)
	})

	t.Run("middleware-propagates-errors", func(t *testing.T) {
		testError := assert.AnError
		middleware := func(next HttpHandlerFunc) HttpHandlerFunc {
			return func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
				resp, err := next(ctx, req)
				if err != nil {
					// Middleware can handle or propagate the error
					return resp, err
				}
				return resp, nil
			}
		}

		chain := HttpMiddlewareChain(middleware)

		handler := func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
			return events.APIGatewayV2HTTPResponse{
				StatusCode: 500,
				Body:       "error occurred",
			}, testError
		}

		wrapped := chain(handler)
		ctx := context.Background()
		req := events.APIGatewayV2HTTPRequest{
			RouteKey: "GET /test",
		}

		resp, err := wrapped(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, testError, err)
		assert.Equal(t, 500, resp.StatusCode)
		assert.Equal(t, "error occurred", resp.Body)
	})

	t.Run("nested-middleware-chain", func(t *testing.T) {
		executionOrder := []string{}

		innerMiddleware := func(next HttpHandlerFunc) HttpHandlerFunc {
			return func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
				executionOrder = append(executionOrder, "inner")
				return next(ctx, req)
			}
		}

		outerMiddleware := func(next HttpHandlerFunc) HttpHandlerFunc {
			return func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
				executionOrder = append(executionOrder, "outer")
				return next(ctx, req)
			}
		}

		innerChain := HttpMiddlewareChain(innerMiddleware)
		outerChain := HttpMiddlewareChain(outerMiddleware, innerChain)

		handler := func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
			executionOrder = append(executionOrder, "handler")
			return events.APIGatewayV2HTTPResponse{
				StatusCode: 200,
				Body:       "success",
			}, nil
		}

		wrapped := outerChain(handler)
		ctx := context.Background()
		req := events.APIGatewayV2HTTPRequest{
			RouteKey: "GET /test",
		}

		resp, err := wrapped(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		// Outer should execute first, then inner, then handler
		expectedOrder := []string{"outer", "inner", "handler"}
		assert.Equal(t, expectedOrder, executionOrder)
	})

	t.Run("middleware-with-context", func(t *testing.T) {
		type contextKey string
		const key contextKey = "middleware-value"

		middleware := func(next HttpHandlerFunc) HttpHandlerFunc {
			return func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
				// Add value to context
				ctx = context.WithValue(ctx, key, "middleware-set-value")
				return next(ctx, req)
			}
		}

		chain := HttpMiddlewareChain(middleware)

		handler := func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
			value := ctx.Value(key)
			valueStr := ""
			if value != nil {
				valueStr = value.(string)
			}
			return events.APIGatewayV2HTTPResponse{
				StatusCode: 200,
				Body:       valueStr,
			}, nil
		}

		wrapped := chain(handler)
		ctx := context.Background()
		req := events.APIGatewayV2HTTPRequest{
			RouteKey: "GET /test",
		}

		resp, err := wrapped(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, "middleware-set-value", resp.Body)
	})
}
