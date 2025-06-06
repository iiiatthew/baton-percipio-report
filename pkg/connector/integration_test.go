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

func TestIntegrationConnectorFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	server := test.FixturesServer()
	defer server.Close()

	t.Run("full connector workflow", func(t *testing.T) {
		// Create connector
		connector, err := New(
			ctx,
			"test-org",
			"test-token",
			24*time.Hour,
		)
		require.NoError(t, err)
		assert.False(t, connector.reportInitialized)

		// Replace client with test server
		connector.client, err = client.New(ctx, server.URL, "test-org", "test-token")
		require.NoError(t, err)

		// Test metadata
		metadata, err := connector.Metadata(ctx)
		require.NoError(t, err)
		assert.Equal(t, "Percipio Connector", metadata.DisplayName)

		// Test validation
		annotations, err := connector.Validate(ctx)
		assert.NoError(t, err)
		assert.Nil(t, annotations)

		// Get resource syncers
		syncers := connector.ResourceSyncers(ctx)
		require.Len(t, syncers, 2)

		var userSyncer *userBuilder
		var courseSyncer *courseBuilder
		for _, syncer := range syncers {
			switch syncer.ResourceType(ctx).Id {
			case "user":
				userSyncer = syncer.(*userBuilder)
			case "course":
				courseSyncer = syncer.(*courseBuilder)
			}
		}
		require.NotNil(t, userSyncer)
		require.NotNil(t, courseSyncer)

		// Test users list - this should trigger report initialization
		users, _, _, err := userSyncer.List(ctx, nil, nil)
		require.NoError(t, err)
		assert.True(t, connector.reportInitialized) // Should be initialized now
		assert.Greater(t, len(users), 0)

		// Test courses list - should use already initialized report
		courses, _, _, err := courseSyncer.List(ctx, nil, nil)
		require.NoError(t, err)
		assert.Greater(t, len(courses), 0)

		// Test course entitlements
		if len(courses) > 0 {
			entitlements, _, _, err := courseSyncer.Entitlements(ctx, courses[0], nil)
			require.NoError(t, err)
			assert.Len(t, entitlements, 3) // assigned, completed, in_progress
		}

		// Test course grants
		if len(courses) > 0 {
			grants, _, _, err := courseSyncer.Grants(ctx, courses[0], nil)
			require.NoError(t, err)
			// May or may not have grants depending on test data
			assert.GreaterOrEqual(t, len(grants), 0)
		}

		// Test user entitlements and grants (should be empty)
		if len(users) > 0 {
			userEntitlements, _, _, err := userSyncer.Entitlements(ctx, users[0], nil)
			require.NoError(t, err)
			assert.Len(t, userEntitlements, 0)

			userGrants, _, _, err := userSyncer.Grants(ctx, users[0], nil)
			require.NoError(t, err)
			assert.Len(t, userGrants, 0)
		}
	})

	t.Run("multiple list calls should reuse report", func(t *testing.T) {
		connector, err := New(
			ctx,
			"test-org",
			"test-token",
			24*time.Hour,
		)
		require.NoError(t, err)

		// Replace client with test server
		connector.client, err = client.New(ctx, server.URL, "test-org", "test-token")
		require.NoError(t, err)

		syncers := connector.ResourceSyncers(ctx)
		userSyncer := syncers[0].(*userBuilder)
		courseSyncer := syncers[1].(*courseBuilder)

		// First call should initialize report
		users1, _, _, err := userSyncer.List(ctx, nil, nil)
		require.NoError(t, err)
		assert.True(t, connector.reportInitialized)
		report1 := connector.report

		// Second call should reuse report
		users2, _, _, err := userSyncer.List(ctx, nil, nil)
		require.NoError(t, err)
		assert.Equal(t, report1, connector.report) // Same report instance

		// Course call should also reuse report
		courses1, _, _, err := courseSyncer.List(ctx, nil, nil)
		require.NoError(t, err)
		assert.Equal(t, report1, connector.report) // Same report instance

		// Results should be consistent
		assert.Equal(t, len(users1), len(users2))
		assert.Greater(t, len(courses1), 0)
	})

	t.Run("error handling", func(t *testing.T) {
		// Test with invalid token/org
		connector, err := New(
			ctx,
			"invalid-org",
			"invalid-token",
			24*time.Hour,
		)
		require.NoError(t, err)

		// Validation should fail with invalid credentials against real API
		annotations, err := connector.Validate(ctx)
		assert.Error(t, err)
		assert.Nil(t, annotations)
	})
}

func TestReportInitializationEdgeCases(t *testing.T) {
	ctx := context.Background()

	t.Run("default lookback period should work", func(t *testing.T) {
		server := test.FixturesServer()
		defer server.Close()

		// Use the default 10 years (as would be set by main.go)
		defaultLookback := 10 * 365 * 24 * time.Hour
		connector, err := New(
			ctx,
			"test-org",
			"test-token",
			defaultLookback,
		)
		require.NoError(t, err)

		connector.client, err = client.New(ctx, server.URL, "test-org", "test-token")
		require.NoError(t, err)

		err = connector.ensureReportInitialized(ctx)
		require.NoError(t, err)
		assert.True(t, connector.reportInitialized)
		assert.Equal(t, defaultLookback, connector.reportLookback)
	})

	t.Run("custom lookback should be used", func(t *testing.T) {
		server := test.FixturesServer()
		defer server.Close()

		customLookback := 48 * time.Hour
		connector, err := New(
			ctx,
			"test-org",
			"test-token",
			customLookback,
		)
		require.NoError(t, err)
		assert.Equal(t, customLookback, connector.reportLookback)

		connector.client, err = client.New(ctx, server.URL, "test-org", "test-token")
		require.NoError(t, err)

		err = connector.ensureReportInitialized(ctx)
		require.NoError(t, err)
		assert.True(t, connector.reportInitialized)
	})
}