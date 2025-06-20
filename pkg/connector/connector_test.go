package connector

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/iiiatthew/baton-percipio-report/pkg/client"
	"github.com/iiiatthew/baton-percipio-report/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnectorNew(t *testing.T) {
	ctx := context.Background()

	t.Run("should create connector successfully", func(t *testing.T) {
		server := test.FixturesServer()
		defer server.Close()

		connector, err := New(
			ctx,
			"test-org",
			"test-token",
			24*time.Hour,
		)

		require.NoError(t, err)
		assert.NotNil(t, connector)
		assert.NotNil(t, connector.client)
		assert.Equal(t, 24*time.Hour, connector.reportLookback)
		assert.Equal(t, ReportNotStarted, connector.reportState)
		assert.Nil(t, connector.report)
	})

	t.Run("should create connector with empty token", func(t *testing.T) {
		// Note: Token validation happens during API calls, not creation
		connector, err := New(
			ctx,
			"test-org",
			"",
			24*time.Hour,
		)

		assert.NoError(t, err)
		assert.NotNil(t, connector)
	})
}

func TestConnectorResourceSyncers(t *testing.T) {
	ctx := context.Background()
	server := test.FixturesServer()
	defer server.Close()

	connector, err := New(
		ctx,
		"test-org",
		"test-token",
		24*time.Hour,
	)
	require.NoError(t, err)

	syncers := connector.ResourceSyncers(ctx)
	assert.Len(t, syncers, 2)

	// Check that we have user and course builders
	foundUser := false
	foundCourse := false
	for _, syncer := range syncers {
		resourceType := syncer.ResourceType(ctx)
		switch resourceType.Id {
		case "user":
			foundUser = true
		case "course":
			foundCourse = true
		}
	}
	assert.True(t, foundUser, "Should have user syncer")
	assert.True(t, foundCourse, "Should have course syncer")
}

func TestConnectorMetadata(t *testing.T) {
	ctx := context.Background()
	server := test.FixturesServer()
	defer server.Close()

	connector, err := New(
		ctx,
		"test-org",
		"test-token",
		24*time.Hour,
	)
	require.NoError(t, err)

	metadata, err := connector.Metadata(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Percipio Connector", metadata.DisplayName)
	assert.Equal(t, "Connector syncing users from Percipio", metadata.Description)
}

func TestConnectorAsset(t *testing.T) {
	ctx := context.Background()
	server := test.FixturesServer()
	defer server.Close()

	connector, err := New(
		ctx,
		"test-org",
		"test-token",
		24*time.Hour,
	)
	require.NoError(t, err)

	// Asset is not implemented, should return empty values
	name, reader, err := connector.Asset(ctx, nil)
	assert.NoError(t, err)
	assert.Empty(t, name)
	assert.Nil(t, reader)
}

func TestConnectorGenerateReport(t *testing.T) {
	ctx := context.Background()

	t.Run("should generate report on first call", func(t *testing.T) {
		server := test.FixturesServer()
		defer server.Close()

		connector, err := New(
			ctx,
			"test-org",
			"test-token",
			24*time.Hour,
		)
		require.NoError(t, err)

		// Mock the client base URL to use test server
		connector.client, err = client.New(
			ctx,
			server.URL,
			"test-org",
			"test-token",
		)
		require.NoError(t, err)

		assert.Equal(t, ReportNotStarted, connector.reportState)
		assert.Nil(t, connector.report)

		err = connector.generateReport(ctx)
		require.NoError(t, err)

		assert.Equal(t, ReportCompleted, connector.reportState)
		assert.NotNil(t, connector.report)
	})

	t.Run("should not regenerate report on subsequent calls", func(t *testing.T) {
		server := test.FixturesServer()
		defer server.Close()

		connector, err := New(
			ctx,
			"test-org",
			"test-token",
			24*time.Hour,
		)
		require.NoError(t, err)

		// Mock the client
		connector.client, err = client.New(
			ctx,
			server.URL,
			"test-org",
			"test-token",
		)
		require.NoError(t, err)

		// First call
		err = connector.generateReport(ctx)
		require.NoError(t, err)
		firstReport := connector.report

		// Second call should not change the report (already completed)
		err = connector.generateReport(ctx)
		require.NoError(t, err)
		assert.Equal(t, firstReport, connector.report)
		assert.Equal(t, ReportCompleted, connector.reportState)
	})

	t.Run("should handle custom lookback period", func(t *testing.T) {
		server := test.FixturesServer()
		defer server.Close()

		// Test with 30 days (720 hours)
		connector, err := New(
			ctx,
			"test-org",
			"test-token",
			720*time.Hour,
		)
		require.NoError(t, err)

		// Mock the client
		connector.client, err = client.New(
			ctx,
			server.URL,
			"test-org",
			"test-token",
		)
		require.NoError(t, err)

		err = connector.generateReport(ctx)
		require.NoError(t, err)

		assert.Equal(t, ReportCompleted, connector.reportState)
	})
}

func TestConnectorValidate(t *testing.T) {
	ctx := context.Background()

	t.Run("should validate successfully with good credentials", func(t *testing.T) {
		server := test.FixturesServer()
		defer server.Close()

		connector, err := New(
			ctx,
			"test-org",
			"test-token",
			24*time.Hour,
		)
		require.NoError(t, err)

		// Mock the client
		connector.client, err = client.New(
			ctx,
			server.URL,
			"test-org",
			"test-token",
		)
		require.NoError(t, err)

		annotations, err := connector.Validate(ctx)
		assert.NoError(t, err)
		assert.Nil(t, annotations)

		// Validate should have generated the report
		assert.Equal(t, ReportCompleted, connector.reportState)
		assert.NotNil(t, connector.report)
	})

	t.Run("should fail validation with bad credentials", func(t *testing.T) {
		connector, err := New(
			ctx,
			"test-org",
			"bad-token",
			24*time.Hour,
		)
		require.NoError(t, err)

		// Use real API URL which should fail with bad token
		annotations, err := connector.Validate(ctx)
		assert.Error(t, err)
		assert.Nil(t, annotations)
		assert.Equal(t, ReportFailed, connector.reportState)
	})
}

func TestWaitForReport(t *testing.T) {
	ctx := context.Background()

	t.Run("should succeed when report is completed", func(t *testing.T) {
		connector := &Connector{
			reportState: ReportCompleted,
		}

		err := connector.waitForReport(ctx)
		assert.NoError(t, err)
	})

	t.Run("should fail when report generation failed", func(t *testing.T) {
		reportErr := fmt.Errorf("test error")
		connector := &Connector{
			reportState: ReportFailed,
			reportError: reportErr,
		}

		err := connector.waitForReport(ctx)
		assert.Error(t, err)
		assert.Equal(t, reportErr, err)
	})

	t.Run("should fail when report not ready", func(t *testing.T) {
		connector := &Connector{
			reportState: ReportNotStarted,
		}

		err := connector.waitForReport(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "report not ready")
	})
}

func TestValidateGeneratesReportForSyncers(t *testing.T) {
	ctx := context.Background()
	server := test.FixturesServer()
	defer server.Close()

	connector, err := New(
		ctx,
		"test-org",
		"test-token",
		24*time.Hour,
	)
	require.NoError(t, err)

	// Mock the client
	connector.client, err = client.New(
		ctx,
		server.URL,
		"test-org",
		"test-token",
	)
	require.NoError(t, err)

	// Validate should generate the report
	annotations, err := connector.Validate(ctx)
	require.NoError(t, err)
	assert.Nil(t, annotations)

	// Report should be completed and available for syncers
	assert.Equal(t, ReportCompleted, connector.reportState)
	assert.NotNil(t, connector.report)

	// Syncers should be able to wait for and use the report
	err = connector.waitForReport(ctx)
	assert.NoError(t, err)
}
