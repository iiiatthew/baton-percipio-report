package connector

import (
	"context"
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
		assert.False(t, connector.reportInitialized)
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

func TestConnectorEnsureReportInitialized(t *testing.T) {
	ctx := context.Background()

	t.Run("should initialize report on first call", func(t *testing.T) {
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

		assert.False(t, connector.reportInitialized)
		assert.Nil(t, connector.report)

		err = connector.ensureReportInitialized(ctx)
		require.NoError(t, err)

		assert.True(t, connector.reportInitialized)
		assert.NotNil(t, connector.report)
	})

	t.Run("should not reinitialize report on subsequent calls", func(t *testing.T) {
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
		err = connector.ensureReportInitialized(ctx)
		require.NoError(t, err)
		firstReport := connector.report

		// Second call should not change the report
		err = connector.ensureReportInitialized(ctx)
		require.NoError(t, err)
		assert.Equal(t, firstReport, connector.report)
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

		err = connector.ensureReportInitialized(ctx)
		require.NoError(t, err)

		assert.True(t, connector.reportInitialized)
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
	})
}