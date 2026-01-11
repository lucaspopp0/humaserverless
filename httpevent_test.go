package humaserverless

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHttpEventToRequest(t *testing.T) {
	t.Run("converts-basic-get-request", func(t *testing.T) {
		event := events.APIGatewayV2HTTPRequest{
			RouteKey:       "GET /api/users",
			RawPath:        "/api/users",
			RawQueryString: "",
			Body:           "",
			Headers: map[string]string{
				"Content-Type": "application/json",
				"User-Agent":   "test-agent",
			},
			RequestContext: events.APIGatewayV2HTTPRequestContext{
				HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
					Path: "/api/users",
				},
			},
		}

		request, err := HttpEventToRequest(event)

		require.NoError(t, err)
		assert.Equal(t, "GET", request.Method)
		assert.Equal(t, "/api/users", request.URL.Path)
		assert.Equal(t, "application/json", request.Header.Get("Content-Type"))
		assert.Equal(t, "test-agent", request.Header.Get("User-Agent"))
	})

	t.Run("converts-post-request-with-body", func(t *testing.T) {
		event := events.APIGatewayV2HTTPRequest{
			RouteKey:       "POST /api/users",
			RawPath:        "/api/users",
			RawQueryString: "",
			Body:           `{"name": "John Doe", "email": "john@example.com"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			RequestContext: events.APIGatewayV2HTTPRequestContext{
				HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
					Path: "/api/users",
				},
			},
		}

		request, err := HttpEventToRequest(event)

		require.NoError(t, err)
		assert.Equal(t, "POST", request.Method)
		assert.Equal(t, "/api/users", request.URL.Path)
		assert.NotNil(t, request.Body)
	})

	t.Run("strips-stage-prefix-from-path", func(t *testing.T) {
		event := events.APIGatewayV2HTTPRequest{
			RouteKey:       "GET /api/users",
			RawPath:        "/prod/api/users",
			RawQueryString: "",
			Body:           "",
			Headers:        map[string]string{},
			RequestContext: events.APIGatewayV2HTTPRequestContext{
				Stage: "prod",
				HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
					Path: "/api/users",
				},
			},
		}

		request, err := HttpEventToRequest(event)

		require.NoError(t, err)
		assert.Equal(t, "/api/users", request.URL.Path)
	})

	t.Run("handles-query-string-parameters", func(t *testing.T) {
		event := events.APIGatewayV2HTTPRequest{
			RouteKey:       "GET /api/users",
			RawPath:        "/api/users",
			RawQueryString: "limit=10&offset=20",
			Body:           "",
			Headers:        map[string]string{},
			RequestContext: events.APIGatewayV2HTTPRequestContext{
				HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
					Path: "/api/users?limit=10&offset=20",
				},
			},
		}

		request, err := HttpEventToRequest(event)

		require.NoError(t, err)
		assert.Contains(t, request.URL.String(), "?")
	})

	t.Run("handles-multiple-headers", func(t *testing.T) {
		event := events.APIGatewayV2HTTPRequest{
			RouteKey:       "GET /api/users",
			RawPath:        "/api/users",
			RawQueryString: "",
			Body:           "",
			Headers: map[string]string{
				"Content-Type":    "application/json",
				"Authorization":   "Bearer token123",
				"X-Custom-Header": "custom-value",
				"User-Agent":      "Mozilla/5.0",
			},
			RequestContext: events.APIGatewayV2HTTPRequestContext{
				HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
					Path: "/api/users",
				},
			},
		}

		request, err := HttpEventToRequest(event)

		require.NoError(t, err)
		assert.Equal(t, "application/json", request.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer token123", request.Header.Get("Authorization"))
		assert.Equal(t, "custom-value", request.Header.Get("X-Custom-Header"))
		assert.Equal(t, "Mozilla/5.0", request.Header.Get("User-Agent"))
	})

	t.Run("handles-put-request", func(t *testing.T) {
		event := events.APIGatewayV2HTTPRequest{
			RouteKey:       "PUT /api/users/123",
			RawPath:        "/api/users/123",
			RawQueryString: "",
			Body:           `{"name": "Jane Doe"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			RequestContext: events.APIGatewayV2HTTPRequestContext{
				HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
					Path: "/api/users/123",
				},
			},
		}

		request, err := HttpEventToRequest(event)

		require.NoError(t, err)
		assert.Equal(t, "PUT", request.Method)
		assert.Equal(t, "/api/users/123", request.URL.Path)
	})

	t.Run("handles-delete-request", func(t *testing.T) {
		event := events.APIGatewayV2HTTPRequest{
			RouteKey:       "DELETE /api/users/123",
			RawPath:        "/api/users/123",
			RawQueryString: "",
			Body:           "",
			Headers:        map[string]string{},
			RequestContext: events.APIGatewayV2HTTPRequestContext{
				HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
					Path: "/api/users/123",
				},
			},
		}

		request, err := HttpEventToRequest(event)

		require.NoError(t, err)
		assert.Equal(t, "DELETE", request.Method)
		assert.Equal(t, "/api/users/123", request.URL.Path)
	})

	t.Run("handles-patch-request", func(t *testing.T) {
		event := events.APIGatewayV2HTTPRequest{
			RouteKey:       "PATCH /api/users/123",
			RawPath:        "/api/users/123",
			RawQueryString: "",
			Body:           `{"email": "newemail@example.com"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			RequestContext: events.APIGatewayV2HTTPRequestContext{
				HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
					Path: "/api/users/123",
				},
			},
		}

		request, err := HttpEventToRequest(event)

		require.NoError(t, err)
		assert.Equal(t, "PATCH", request.Method)
	})

	t.Run("handles-empty-headers", func(t *testing.T) {
		event := events.APIGatewayV2HTTPRequest{
			RouteKey:       "GET /api/users",
			RawPath:        "/api/users",
			RawQueryString: "",
			Body:           "",
			Headers:        map[string]string{},
			RequestContext: events.APIGatewayV2HTTPRequestContext{
				HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
					Path: "/api/users",
				},
			},
		}

		request, err := HttpEventToRequest(event)

		require.NoError(t, err)
		assert.NotNil(t, request)
		assert.Equal(t, 0, len(request.Header))
	})

	t.Run("handles-root-path", func(t *testing.T) {
		event := events.APIGatewayV2HTTPRequest{
			RouteKey:       "GET /",
			RawPath:        "/",
			RawQueryString: "",
			Body:           "",
			Headers:        map[string]string{},
			RequestContext: events.APIGatewayV2HTTPRequestContext{
				HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
					Path: "/",
				},
			},
		}

		request, err := HttpEventToRequest(event)

		require.NoError(t, err)
		assert.Equal(t, "/", request.URL.Path)
	})

	t.Run("handles-nested-path", func(t *testing.T) {
		event := events.APIGatewayV2HTTPRequest{
			RouteKey:       "GET /api/v1/users/123/posts/456",
			RawPath:        "/api/v1/users/123/posts/456",
			RawQueryString: "",
			Body:           "",
			Headers:        map[string]string{},
			RequestContext: events.APIGatewayV2HTTPRequestContext{
				HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
					Path: "/api/v1/users/123/posts/456",
				},
			},
		}

		request, err := HttpEventToRequest(event)

		require.NoError(t, err)
		assert.Equal(t, "/api/v1/users/123/posts/456", request.URL.Path)
	})
}

func TestResponseToHttpEvent(t *testing.T) {
	t.Run("converts-basic-200-response", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		recorder.WriteHeader(http.StatusOK)
		recorder.WriteString(`{"message": "success"}`)

		event := ResponseToHttpEvent(*recorder)

		assert.Equal(t, http.StatusOK, event.StatusCode)
		assert.Equal(t, `{"message": "success"}`, event.Body)
		assert.NotNil(t, event.Headers)
		assert.NotNil(t, event.MultiValueHeaders)
	})

	t.Run("converts-201-created-response", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		recorder.WriteHeader(http.StatusCreated)
		recorder.WriteString(`{"id": "123"}`)

		event := ResponseToHttpEvent(*recorder)

		assert.Equal(t, http.StatusCreated, event.StatusCode)
		assert.Equal(t, `{"id": "123"}`, event.Body)
	})

	t.Run("converts-204-no-content-response", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		recorder.WriteHeader(http.StatusNoContent)

		event := ResponseToHttpEvent(*recorder)

		assert.Equal(t, http.StatusNoContent, event.StatusCode)
		assert.Equal(t, "", event.Body)
	})

	t.Run("converts-400-bad-request-response", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		recorder.WriteHeader(http.StatusBadRequest)
		recorder.WriteString(`{"error": "invalid input"}`)

		event := ResponseToHttpEvent(*recorder)

		assert.Equal(t, http.StatusBadRequest, event.StatusCode)
		assert.Equal(t, `{"error": "invalid input"}`, event.Body)
	})

	t.Run("converts-500-internal-server-error-response", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		recorder.WriteHeader(http.StatusInternalServerError)
		recorder.WriteString(`{"error": "server error"}`)

		event := ResponseToHttpEvent(*recorder)

		assert.Equal(t, http.StatusInternalServerError, event.StatusCode)
		assert.Equal(t, `{"error": "server error"}`, event.Body)
	})

	t.Run("handles-single-header", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		recorder.Header().Set("Content-Type", "application/json")
		recorder.WriteHeader(http.StatusOK)
		recorder.WriteString(`{}`)

		event := ResponseToHttpEvent(*recorder)

		assert.Equal(t, "application/json", event.Headers["Content-Type"])
		assert.Equal(
			t,
			[]string{"application/json"},
			event.MultiValueHeaders["Content-Type"],
		)
	})

	t.Run("handles-multiple-headers", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		recorder.Header().Set("Content-Type", "application/json")
		recorder.Header().Set("X-Request-Id", "req-123")
		recorder.Header().Set("X-Custom-Header", "custom-value")
		recorder.WriteHeader(http.StatusOK)
		recorder.WriteString(`{}`)

		event := ResponseToHttpEvent(*recorder)

		assert.Equal(t, "application/json", event.Headers["Content-Type"])
		assert.Equal(t, "req-123", event.Headers["X-Request-Id"])
		assert.Equal(t, "custom-value", event.Headers["X-Custom-Header"])
	})

	t.Run("handles-multi-value-headers", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		recorder.Header().Add("Set-Cookie", "session=abc123")
		recorder.Header().Add("Set-Cookie", "token=xyz789")
		recorder.Header().Add("Set-Cookie", "preferences=dark-mode")
		recorder.WriteHeader(http.StatusOK)
		recorder.WriteString(`{}`)

		event := ResponseToHttpEvent(*recorder)

		// Should use the last value for the single-value header
		assert.Equal(t, "preferences=dark-mode", event.Headers["Set-Cookie"])

		// Should include all values in the multi-value header
		assert.Equal(
			t,
			[]string{"session=abc123", "token=xyz789", "preferences=dark-mode"},
			event.MultiValueHeaders["Set-Cookie"],
		)
	})

	t.Run("handles-empty-headers", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		recorder.WriteHeader(http.StatusOK)
		recorder.WriteString(`{}`)

		event := ResponseToHttpEvent(*recorder)

		assert.NotNil(t, event.Headers)
		assert.NotNil(t, event.MultiValueHeaders)
	})

	t.Run("handles-empty-body", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		recorder.WriteHeader(http.StatusOK)

		event := ResponseToHttpEvent(*recorder)

		assert.Equal(t, "", event.Body)
	})

	t.Run("handles-large-body", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		largeBody := `{"data": "` + string(make([]byte, 10000)) + `"}`
		recorder.WriteHeader(http.StatusOK)
		recorder.WriteString(largeBody)

		event := ResponseToHttpEvent(*recorder)

		assert.Equal(t, largeBody, event.Body)
	})

	t.Run("handles-404-not-found-response", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		recorder.WriteHeader(http.StatusNotFound)
		recorder.WriteString(`{"error": "not found"}`)

		event := ResponseToHttpEvent(*recorder)

		assert.Equal(t, http.StatusNotFound, event.StatusCode)
		assert.Equal(t, `{"error": "not found"}`, event.Body)
	})

	t.Run("handles-redirect-response", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		recorder.Header().Set("Location", "/new-location")
		recorder.WriteHeader(http.StatusMovedPermanently)

		event := ResponseToHttpEvent(*recorder)

		assert.Equal(t, http.StatusMovedPermanently, event.StatusCode)
		assert.Equal(t, "/new-location", event.Headers["Location"])
	})

	t.Run("preserves-header-case", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		recorder.Header().Set("X-Custom-Header", "value")
		recorder.WriteHeader(http.StatusOK)

		event := ResponseToHttpEvent(*recorder)

		assert.Contains(t, event.Headers, "X-Custom-Header")
	})
}
