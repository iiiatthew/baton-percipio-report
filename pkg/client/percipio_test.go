package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientNew(t *testing.T) {
	ctx := context.Background()

	t.Run("should create client successfully", func(t *testing.T) {
		client, err := New(
			ctx,
			"https://api.example.com",
			"test-org",
			"test-token",
		)

		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "test-org", client.organizationId)
		assert.Equal(t, "test-token", client.bearerToken)
		assert.NotNil(t, client.StatusesStore)
		assert.Nil(t, client.loadedReport)
	})

	t.Run("should fail with invalid URL", func(t *testing.T) {
		_, err := New(
			ctx,
			"://invalid-url", // This should definitely be invalid
			"test-org",
			"test-token",
		)

		assert.Error(t, err)
	})
}

func TestGenerateLearningActivityReport(t *testing.T) {
	ctx := context.Background()

	t.Run("should generate report successfully", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Contains(t, r.URL.Path, "/reporting/v1/organizations/test-org/report-requests/learning-activity")
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id": "report-123", "status": "PENDING"}`))
		}))
		defer server.Close()

		client, err := New(ctx, server.URL, "test-org", "test-token")
		require.NoError(t, err)

		rateLimit, err := client.GenerateLearningActivityReport(ctx, 24*time.Hour)
		assert.NoError(t, err)
		assert.NotNil(t, rateLimit)
		assert.Equal(t, "report-123", client.ReportStatus.Id)
		assert.Equal(t, "PENDING", client.ReportStatus.Status)
	})

	t.Run("should handle custom lookback period", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id": "report-123", "status": "PENDING"}`))
		}))
		defer server.Close()

		client, err := New(ctx, server.URL, "test-org", "test-token")
		require.NoError(t, err)

		// Test with 30 days (720 hours)
		rateLimit, err := client.GenerateLearningActivityReport(ctx, 720*time.Hour)
		assert.NoError(t, err)
		assert.NotNil(t, rateLimit)
	})

	t.Run("should handle API error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error": "unauthorized"}`))
		}))
		defer server.Close()

		client, err := New(ctx, server.URL, "test-org", "test-token")
		require.NoError(t, err)

		_, err = client.GenerateLearningActivityReport(ctx, 24*time.Hour)
		assert.Error(t, err)
	})
}

func TestGetLearningActivityReport(t *testing.T) {
	ctx := context.Background()

	t.Run("should get report when ready immediately", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.URL.Path, "/reporting/v1/organizations/test-org/report-requests/report-123")

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			// Return report data immediately (no polling simulation)
			_, _ = w.Write([]byte(`[
				{
					"userId": "user1",
					"firstName": "John",
					"lastName": "Doe",
					"emailAddress": "john@example.com",
					"contentId": "course1",
					"contentTitle": "Test Course",
					"status": "Completed"
				}
			]`))
		}))
		defer server.Close()

		client, err := New(ctx, server.URL, "test-org", "test-token")
		require.NoError(t, err)
		client.ReportStatus = ReportStatus{Id: "report-123", Status: "PENDING"}

		rateLimit, err := client.GetLearningActivityReport(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, rateLimit)
		assert.Equal(t, "done", client.ReportStatus.Status)
		assert.NotNil(t, client.loadedReport)
		assert.Len(t, *client.loadedReport, 1)
	})

	t.Run("should handle failed report", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id": "report-123", "status": "FAILED"}`))
		}))
		defer server.Close()

		client, err := New(ctx, server.URL, "test-org", "test-token")
		require.NoError(t, err)
		client.ReportStatus = ReportStatus{Id: "report-123", Status: "PENDING"}

		_, err = client.GetLearningActivityReport(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "report generation failed")
	})
}

func TestGetLoadedReport(t *testing.T) {
	ctx := context.Background()

	client, err := New(ctx, "https://api.example.com", "test-org", "test-token")
	require.NoError(t, err)

	// Initially should be nil
	assert.Nil(t, client.GetLoadedReport())

	// Set a report
	testReport := &Report{
		{UserId: "michael.bolton@initech.com", ContentId: "bs_adg02_a23_enus"},
	}
	client.loadedReport = testReport

	// Should return the report
	assert.Equal(t, testReport, client.GetLoadedReport())
}
