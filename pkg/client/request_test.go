package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientGetUrl(t *testing.T) {
	client := &Client{
		baseUrl:        mustParseURL("https://api.example.com"),
		organizationId: "test-org",
	}

	t.Run("should build URL with path only", func(t *testing.T) {
		url := client.getUrl("/test/path/%s", nil)
		assert.Equal(t, "https://api.example.com/test/path/test-org", url.String())
	})

	t.Run("should build URL with query parameters", func(t *testing.T) {
		params := map[string]any{
			"limit":  10,
			"offset": 20,
			"active": true,
			"name":   "test",
		}
		url := client.getUrl("/test/path/%s", params)

		assert.Equal(t, "test/path/test-org", url.Path)

		// Check query parameters
		query := url.Query()
		assert.Equal(t, "10", query.Get("limit"))
		assert.Equal(t, "20", query.Get("offset"))
		assert.Equal(t, "true", query.Get("active"))
		assert.Equal(t, "test", query.Get("name"))
	})

	t.Run("should ignore unsupported parameter types", func(t *testing.T) {
		params := map[string]any{
			"valid":   "test",
			"invalid": []string{"unsupported", "type"},
		}
		url := client.getUrl("/test/path/%s", params)

		query := url.Query()
		assert.Equal(t, "test", query.Get("valid"))
		assert.Empty(t, query.Get("invalid"))
	})
}

func TestWithBearerToken(t *testing.T) {
	// This is mainly testing that the function exists and creates a valid option
	// The actual functionality is tested in integration tests
	option := WithBearerToken("test-token")
	assert.NotNil(t, option)
}

func TestClientDoRequest(t *testing.T) {
	ctx := context.Background()

	t.Run("should make successful GET request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
			assert.Equal(t, "application/json", r.Header.Get("Accept"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"result": "success"}`))
		}))
		defer server.Close()

		client, err := New(ctx, server.URL, "test-org", "test-token")
		require.NoError(t, err)

		var target map[string]interface{}
		response, rateLimit, err := client.doRequest(
			ctx,
			"GET",
			"/test",
			nil,
			nil,
			&target,
		)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, rateLimit)
		assert.Equal(t, "success", target["result"])
		assert.Equal(t, http.StatusOK, response.StatusCode)
	})

	t.Run("should make successful POST request with body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id": "123"}`))
		}))
		defer server.Close()

		client, err := New(ctx, server.URL, "test-org", "test-token")
		require.NoError(t, err)

		payload := map[string]interface{}{
			"name": "test",
		}
		var target map[string]interface{}
		response, rateLimit, err := client.doRequest(
			ctx,
			"POST",
			"/test",
			nil,
			payload,
			&target,
		)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, rateLimit)
		assert.Equal(t, "123", target["id"])
		assert.Equal(t, http.StatusCreated, response.StatusCode)
	})

	t.Run("should handle API error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error": "bad request"}`))
		}))
		defer server.Close()

		client, err := New(ctx, server.URL, "test-org", "test-token")
		require.NoError(t, err)

		var target map[string]interface{}
		_, _, err = client.doRequest(
			ctx,
			"GET",
			"/test",
			nil,
			nil,
			&target,
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error making GET request")
	})

	t.Run("should handle query parameters", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "10", r.URL.Query().Get("limit"))
			assert.Equal(t, "test", r.URL.Query().Get("filter"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}))
		defer server.Close()

		client, err := New(ctx, server.URL, "test-org", "test-token")
		require.NoError(t, err)

		queryParams := map[string]any{
			"limit":  10,
			"filter": "test",
		}
		var target map[string]interface{}
		_, _, err = client.doRequest(
			ctx,
			"GET",
			"/test",
			queryParams,
			nil,
			&target,
		)

		assert.NoError(t, err)
	})
}

func TestClientGetAndPost(t *testing.T) {
	ctx := context.Background()

	t.Run("get method should work", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"test": "value"}`))
		}))
		defer server.Close()

		client, err := New(ctx, server.URL, "test-org", "test-token")
		require.NoError(t, err)

		var target map[string]interface{}
		response, rateLimit, err := client.get(ctx, "/test", nil, &target)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, rateLimit)
		assert.Equal(t, "value", target["test"])
	})

	t.Run("post method should work", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"created": "success"}`))
		}))
		defer server.Close()

		client, err := New(ctx, server.URL, "test-org", "test-token")
		require.NoError(t, err)

		body := map[string]interface{}{"name": "test"}
		var target map[string]interface{}
		response, rateLimit, err := client.post(ctx, "/test", body, &target)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, rateLimit)
		assert.Equal(t, "success", target["created"])
	})
}

// Helper function for tests
func mustParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return u
}
