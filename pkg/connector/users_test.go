package connector

import (
	"context"
	"testing"
	"time"

	"github.com/iiiatthew/baton-percipio-report/pkg/client"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/iiiatthew/baton-percipio-report/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsersList(t *testing.T) {
	ctx := context.Background()

	t.Run("should get users from report data", func(t *testing.T) {
		// Create a mock connector with report data
		connector := &Connector{
			reportInitialized: true,
			report: &client.Report{
				{
					UserUUID:      "a77840ca-ea10-4da8-b64f-bddf714c47a0",
					FirstName:     "John",
					LastName:      "Doe",
					EmailAddress:  "john@example.com",
					ContentUUID:   "1a3a3f54-b601-4d45-a234-038c980ee20f",
					Status:        "Completed",
					CompletedDate: "2023-01-15",
				},
				{
					UserUUID:     "a77840ca-ea10-4da8-b64f-bddf714c47a1",
					FirstName:    "Jane",
					LastName:     "Smith",
					EmailAddress: "jane@example.com",
					ContentUUID:  "1a3a3f54-b601-4d45-a234-038c980ee20f",
					Status:       "Started",
					FirstAccess:  "2023-01-10",
				},
			},
		}

		c := newUserBuilder(nil, nil, connector)

		resources, nextToken, annotations, err := c.List(ctx, nil, &pagination.Token{})

		require.NoError(t, err)
		assert.Empty(t, nextToken) // No pagination for report-based data
		test.AssertNoRatelimitAnnotations(t, annotations)
		require.Len(t, resources, 2) // 2 unique users

		// Check user1
		user1 := findResourceById(resources, "a77840ca-ea10-4da8-b64f-bddf714c47a0")
		require.NotNil(t, user1)
		assert.Equal(t, "John Doe", user1.DisplayName)

		// Check user2
		user2 := findResourceById(resources, "a77840ca-ea10-4da8-b64f-bddf714c47a1")
		require.NotNil(t, user2)
		assert.Equal(t, "Jane Smith", user2.DisplayName)
	})

	t.Run("should handle duplicate users with date priority", func(t *testing.T) {
		connector := &Connector{
			reportInitialized: true,
			report: &client.Report{
				{
					UserUUID:     "a77840ca-ea10-4da8-b64f-bddf714c47a0",
					FirstName:    "John",
					LastName:     "Doe",
					EmailAddress: "john@example.com",
					ContentUUID:  "1a3a3f54-b601-4d45-a234-038c980ee20f",
					Status:       "Started",
					FirstAccess:  "2023-01-10",
				},
				{
					UserUUID:      "a77840ca-ea10-4da8-b64f-bddf714c47a0",
					FirstName:     "Johnny", // Updated name
					LastName:      "Doe",
					EmailAddress:  "johnny@example.com", // Updated email
					ContentUUID:   "1a3a3f54-b601-4d45-a234-038c980ee20f",
					Status:        "Completed",
					CompletedDate: "2023-01-15", // More recent
				},
			},
		}

		c := newUserBuilder(nil, nil, connector)

		resources, _, _, err := c.List(ctx, nil, &pagination.Token{})

		require.NoError(t, err)
		require.Len(t, resources, 1) // Should deduplicate to 1 user

		user := resources[0]
		assert.Equal(t, "a77840ca-ea10-4da8-b64f-bddf714c47a0", user.Id.Resource)
		assert.Equal(t, "Johnny Doe", user.DisplayName) // Should use more recent data
	})

	t.Run("should handle missing email gracefully", func(t *testing.T) {
		connector := &Connector{
			reportInitialized: true,
			report: &client.Report{
				{
					UserUUID:  "a77840ca-ea10-4da8-b64f-bddf714c47a0",
					FirstName: "John",
					LastName:  "Doe",
					// EmailAddress missing
					ContentUUID: "1a3a3f54-b601-4d45-a234-038c980ee20f",
					Status:      "Completed",
				},
			},
		}

		c := newUserBuilder(nil, nil, connector)

		resources, _, _, err := c.List(ctx, nil, &pagination.Token{})

		require.NoError(t, err)
		require.Len(t, resources, 1)
		// Should still create user even without email
	})

	t.Run("should handle empty report", func(t *testing.T) {
		connector := &Connector{
			reportInitialized: true,
			report:            &client.Report{},
		}

		c := newUserBuilder(nil, nil, connector)

		resources, _, _, err := c.List(ctx, nil, &pagination.Token{})

		require.NoError(t, err)
		assert.Len(t, resources, 0)
	})

	t.Run("should initialize report if not done", func(t *testing.T) {
		server := test.FixturesServer()
		defer server.Close()

		percipioClient, err := client.New(ctx, server.URL, "mock", "token")
		require.NoError(t, err)

		connector := &Connector{
			client:            percipioClient,
			reportLookback:    24 * time.Hour,
			reportInitialized: false,
		}

		c := newUserBuilder(percipioClient, nil, connector)

		resources, _, _, err := c.List(ctx, nil, &pagination.Token{})

		require.NoError(t, err)
		assert.True(t, connector.reportInitialized)
		// Should have users from test fixtures
		assert.Greater(t, len(resources), 0)
	})
}

func TestUserEntitlements(t *testing.T) {
	ctx := context.Background()
	c := newUserBuilder(nil, nil, nil)

	entitlements, nextToken, annotations, err := c.Entitlements(ctx, nil, nil)

	assert.NoError(t, err)
	assert.Nil(t, entitlements)
	assert.Empty(t, nextToken)
	assert.Nil(t, annotations)
}

func TestUserGrants(t *testing.T) {
	ctx := context.Background()
	c := newUserBuilder(nil, nil, nil)

	grants, nextToken, annotations, err := c.Grants(ctx, nil, nil)

	assert.NoError(t, err)
	assert.Nil(t, grants)
	assert.Empty(t, nextToken)
	assert.Nil(t, annotations)
}

func TestGetDisplayName(t *testing.T) {
	testCases := []struct {
		user     client.User
		expected string
		desc     string
	}{
		{
			client.User{FirstName: "John", LastName: "Doe", Email: ""},
			"John Doe",
			"full name",
		},
		{
			client.User{FirstName: "John", LastName: "", Email: ""},
			"John",
			"first name only",
		},
		{
			client.User{FirstName: "", LastName: "Doe", Email: ""},
			"Doe",
			"last name only",
		},
		{
			client.User{Email: "john@example.com", FirstName: "", LastName: "", LoginID: ""},
			"john@example.com",
			"email fallback",
		},
		{
			client.User{LoginID: "user123", FirstName: "", LastName: "", Email: ""},
			"user123",
			"id fallback",
		},
		{
			client.User{FirstName: "", LastName: "", Email: "", LoginID: ""},
			"<no name>",
			"default fallback",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := getDisplayName(tc.user)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// Helper function to find resource by ID
func findResourceById(resources []*v2.Resource, id string) *v2.Resource {
	for _, resource := range resources {
		if resource.Id.Resource == id {
			return resource
		}
	}
	return nil
}
