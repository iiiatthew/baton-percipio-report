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
		connector := &Connector{
			reportState: ReportCompleted,
			report: &client.Report{
				{
					UserId:       "michael.bolton@initech.com",
					FirstName:    "Michael",
					LastName:     "Bolton",
					EmailAddress: "michael.bolton@initech.com",
					ContentId:    "bs_adg02_a23_enus",
					ContentTitle: "Case Studies: Successful Data Privacy Implementations",
					ContentType:  "Course",
					Status:       "Completed",
				},
				{
					UserId:       "milton.waddams@initech.com",
					FirstName:    "Milton",
					LastName:     "Waddams",
					EmailAddress: "milton.waddams@initech.com",
					ContentId:    "bs_adg02_a23_enus",
					ContentTitle: "Case Studies: Successful Data Privacy Implementations",
					ContentType:  "Course",
					Status:       "Started",
				},
				{
					UserId:       "michael.bolton@initech.com",
					FirstName:    "Michael",
					LastName:     "Bolton",
					EmailAddress: "michael.bolton@initech.com",
					ContentId:    "another_course_id",
					ContentTitle: "Advanced Go Patterns",
					ContentType:  "Assessment",
					Status:       "Active",
				},
				{
					UserId:       "peter.gibbons@initech.com",
					FirstName:    "Peter",
					LastName:     "Gibbons",
					EmailAddress: "peter.gibbons@initech.com",
					ContentId:    "bs_adg02_a23_enus",
					ContentTitle: "Case Studies: Successful Data Privacy Implementations",
					ContentType:  "Course",
					Status:       "",
				},
				{
					UserId:       "bill.lumbergh@initech.com",
					FirstName:    "Bill",
					LastName:     "Lumbergh",
					EmailAddress: "bill.lumbergh@initech.com",
					ContentId:    "another_course_id",
					ContentTitle: "Advanced Go Patterns",
					ContentType:  "Assessment",
					Status:       "UnknownStatus",
				},
			},
		}

		u := newUserBuilder(nil, nil, connector)

		resources, nextToken, annotations, err := u.List(ctx, nil, &pagination.Token{})

		require.NoError(t, err)
		assert.Empty(t, nextToken)
		test.AssertNoRatelimitAnnotations(t, annotations)
		require.Len(t, resources, 4)

		userByID := make(map[string]*v2.Resource)
		for _, user := range resources {
			userByID[user.Id.Resource] = user
		}

		// Verify Michael
		michael := userByID["michael.bolton@initech.com"]
		require.NotNil(t, michael)
		assert.Equal(t, "Michael Bolton", michael.DisplayName)

		// Verify Milton
		milton := userByID["milton.waddams@initech.com"]
		require.NotNil(t, milton)
		assert.Equal(t, "Milton Waddams", milton.DisplayName)

		// Verify Peter
		peter := userByID["peter.gibbons@initech.com"]
		require.NotNil(t, peter)
		assert.Equal(t, "Peter Gibbons", peter.DisplayName)

		// Verify Bill
		bill := userByID["bill.lumbergh@initech.com"]
		require.NotNil(t, bill)
		assert.Equal(t, "Bill Lumbergh", bill.DisplayName)
	})

	t.Run("should wait for report that was generated during validation", func(t *testing.T) {
		server := test.FixturesServer()
		defer server.Close()

		percipioClient, err := client.New(ctx, server.URL, "mock", "token")
		require.NoError(t, err)

		connector := &Connector{
			client:         percipioClient,
			reportLookback: 24 * time.Hour,
			reportState:    ReportCompleted,
			report: &client.Report{
				{
					UserId:       "test@example.com",
					FirstName:    "Test",
					LastName:     "User",
					EmailAddress: "test@example.com",
					ContentId:    "test-course",
					ContentTitle: "Test Course",
					ContentType:  "Course",
					Status:       "Completed",
				},
			},
		}

		u := newUserBuilder(percipioClient, nil, connector)

		resources, _, _, err := u.List(ctx, nil, &pagination.Token{})

		require.NoError(t, err)
		assert.Equal(t, ReportCompleted, connector.reportState)
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
		name     string
		user     client.User
		expected string
	}{
		{
			name: "full name",
			user: client.User{
				FirstName: "Michael",
				LastName:  "Bolton",
			},
			expected: "Michael Bolton",
		},
		{
			name: "first name only",
			user: client.User{
				FirstName: "Michael",
				LastName:  "",
			},
			expected: "Michael ",
		},
		{
			name: "last name only",
			user: client.User{
				FirstName: "",
				LastName:  "Bolton",
			},
			expected: " Bolton",
		},
		{
			name: "empty names",
			user: client.User{
				FirstName: "",
				LastName:  "",
			},
			expected: " ",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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

func TestUsersGrants(t *testing.T) {
	ctx := context.Background()

	t.Run("should return empty grants for user", func(t *testing.T) {
		u := newUserBuilder(nil, nil, nil)
		user := &v2.Resource{
			DisplayName: "Michael Bolton",
			Id: &v2.ResourceId{
				ResourceType: "user",
				Resource:     "michael.bolton@initech.com",
			},
		}

		grants, nextToken, annotations, err := u.Grants(ctx, user, &pagination.Token{})

		require.NoError(t, err)
		assert.Empty(t, nextToken)
		assert.Nil(t, annotations)
		assert.Len(t, grants, 0)
	})
}

func TestUserResource(t *testing.T) {
	t.Run("should create user resource with display name", func(t *testing.T) {
		user := client.User{
			Id:        "michael.bolton@initech.com",
			Email:     "michael.bolton@initech.com",
			FirstName: "Michael",
			LastName:  "Bolton",
		}

		resource, err := userResource(user, nil)

		require.NoError(t, err)
		assert.Equal(t, "Michael Bolton", resource.DisplayName)
		assert.Equal(t, "michael.bolton@initech.com", resource.Id.Resource)
		assert.Equal(t, "user", resource.Id.ResourceType)
	})

	t.Run("should handle empty names", func(t *testing.T) {
		user := client.User{
			Id:        "michael.bolton@initech.com",
			Email:     "michael.bolton@initech.com",
			FirstName: "",
			LastName:  "",
		}

		resource, err := userResource(user, nil)

		require.NoError(t, err)
		assert.Equal(t, " ", resource.DisplayName)
	})
}
