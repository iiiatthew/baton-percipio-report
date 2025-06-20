package connector

import (
	"context"
	"strings"
	"testing"
	"time"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/iiiatthew/baton-percipio-report/pkg/client"
	"github.com/iiiatthew/baton-percipio-report/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCoursesList(t *testing.T) {
	ctx := context.Background()

	t.Run("should get courses from report data", func(t *testing.T) {
		connector := &Connector{
			reportInitialized: true,
			report: &client.Report{
				{
					UserUUID:     "a77840ca-ea10-4da8-b64f-bddf714c47a0",
					ContentUUID:  "1a3a3f54-b601-4d45-a234-038c980ee20f",
					ContentTitle: "Introduction to Go",
					Status:       "Completed",
				},
				{
					UserUUID:     "a77840ca-ea10-4da8-b64f-bddf714c47a1",
					ContentUUID:  "1a3a3f54-b601-4d45-a234-038c980ee20f", // Same course
					ContentTitle: "Introduction to Go",
					Status:       "Started",
				},
				{
					UserUUID:     "a77840ca-ea10-4da8-b64f-bddf714c47a0",
					ContentUUID:  "1a3a3f54-b601-4d45-a234-038c980ee20f",
					ContentTitle: "Advanced Go Patterns",
					Status:       "Active",
				},
			},
		}

		c := newCourseBuilder(nil, nil, connector)

		resources, nextToken, annotations, err := c.List(ctx, nil, &pagination.Token{})

		require.NoError(t, err)
		assert.Empty(t, nextToken) // No pagination for report-based data
		test.AssertNoRatelimitAnnotations(t, annotations)
		require.Len(t, resources, 2) // 2 unique courses

		// Check course1
		course1 := findResourceById(resources, "course1")
		require.NotNil(t, course1)
		assert.Equal(t, "Introduction to Go", course1.DisplayName)

		// Check course2
		course2 := findResourceById(resources, "course2")
		require.NotNil(t, course2)
		assert.Equal(t, "Advanced Go Patterns", course2.DisplayName)
	})

	t.Run("should handle missing contentId", func(t *testing.T) {
		connector := &Connector{
			reportInitialized: true,
			report: &client.Report{
				{
					UserUUID:     "a77840ca-ea10-4da8-b64f-bddf714c47a0",
					ContentUUID:  "", // Missing contentUUID
					ContentTitle: "Course Without ID",
					Status:       "Completed",
				},
				{
					UserUUID:     "a77840ca-ea10-4da8-b64f-bddf714c47a0",
					ContentUUID:  "1a3a3f54-b601-4d45-a234-038c980ee20f",
					ContentTitle: "Valid Course",
					Status:       "Completed",
				},
			},
		}

		c := newCourseBuilder(nil, nil, connector)

		resources, _, _, err := c.List(ctx, nil, &pagination.Token{})

		require.NoError(t, err)
		require.Len(t, resources, 1) // Should skip course without ID
		assert.Equal(t, "course1", resources[0].Id.Resource)
	})

	t.Run("should handle missing title", func(t *testing.T) {
		connector := &Connector{
			reportInitialized: true,
			report: &client.Report{
				{
					UserUUID:     "a77840ca-ea10-4da8-b64f-bddf714c47a0",
					ContentUUID:  "1a3a3f54-b601-4d45-a234-038c980ee20f",
					ContentTitle: "", // Missing title
					Status:       "Completed",
				},
			},
		}

		c := newCourseBuilder(nil, nil, connector)

		resources, _, _, err := c.List(ctx, nil, &pagination.Token{})

		require.NoError(t, err)
		require.Len(t, resources, 1)
		// Should use course ID as display name when title is missing
		assert.Equal(t, "course1", resources[0].DisplayName)
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

		c := newCourseBuilder(percipioClient, nil, connector)

		resources, _, _, err := c.List(ctx, nil, &pagination.Token{})

		require.NoError(t, err)
		assert.True(t, connector.reportInitialized)
		// Should have courses from test fixtures
		assert.Greater(t, len(resources), 0)
	})
}

func TestCoursesEntitlements(t *testing.T) {
	ctx := context.Background()

	c := newCourseBuilder(nil, nil, nil)
	course := &v2.Resource{
		DisplayName: "Test Course",
		Id: &v2.ResourceId{
			ResourceType: "course",
			Resource:     "course1",
		},
	}

	entitlements, nextToken, annotations, err := c.Entitlements(ctx, course, &pagination.Token{})

	require.NoError(t, err)
	assert.Empty(t, nextToken)
	assert.Nil(t, annotations)
	require.Len(t, entitlements, 3) // assigned, completed, in_progress

	// Check entitlement types
	entitlementSlugs := make([]string, len(entitlements))
	for i, ent := range entitlements {
		entitlementSlugs[i] = ent.Slug
	}
	assert.Contains(t, entitlementSlugs, "assigned")
	assert.Contains(t, entitlementSlugs, "completed")
	assert.Contains(t, entitlementSlugs, "in_progress")
}

func TestCoursesGrants(t *testing.T) {
	ctx := context.Background()

	t.Run("should return grants for course", func(t *testing.T) {
		// Set up status store
		statusStore := make(client.StatusesStore)
		statusStore["course1"] = map[string]string{
			"user1": "completed",
			"user2": "in_progress",
		}

		percipioClient := &client.Client{
			StatusesStore: statusStore,
		}

		c := newCourseBuilder(percipioClient, nil, nil)
		course := &v2.Resource{
			DisplayName: "Test Course",
			Id: &v2.ResourceId{
				ResourceType: "course",
				Resource:     "course1",
			},
		}

		grants, nextToken, annotations, err := c.Grants(ctx, course, &pagination.Token{})

		require.NoError(t, err)
		assert.Empty(t, nextToken)
		test.AssertNoRatelimitAnnotations(t, annotations)
		require.Len(t, grants, 2)

		// Check grants - extract entitlement name from entitlement ID
		// Entitlement ID format is "resourceType:resourceId:entitlementName"
		grantsByUser := make(map[string]string)
		for _, grant := range grants {
			entitlementId := grant.Entitlement.Id
			// Split by ":" and get the last part (entitlement name)
			parts := strings.Split(entitlementId, ":")
			if len(parts) >= 3 {
				entitlementName := parts[len(parts)-1]
				grantsByUser[grant.Principal.Id.Resource] = entitlementName
			}
		}
		assert.Equal(t, "completed", grantsByUser["user1"])
		assert.Equal(t, "in_progress", grantsByUser["user2"])
	})

	t.Run("should handle course with no grants", func(t *testing.T) {
		percipioClient := &client.Client{
			StatusesStore: make(client.StatusesStore),
		}

		c := newCourseBuilder(percipioClient, nil, nil)
		course := &v2.Resource{
			DisplayName: "Empty Course",
			Id: &v2.ResourceId{
				ResourceType: "course",
				Resource:     "empty-course",
			},
		}

		grants, _, _, err := c.Grants(ctx, course, &pagination.Token{})

		require.NoError(t, err)
		assert.Len(t, grants, 0)
	})
}

func TestCourseResource(t *testing.T) {
	t.Run("should create course resource with name", func(t *testing.T) {
		course := client.Course{
			Id:          "course1",
			CourseTitle: "Test Course",
		}

		resource, err := courseResource(course, nil)

		require.NoError(t, err)
		assert.Equal(t, "Test Course", resource.DisplayName)
		assert.Equal(t, "course1", resource.Id.Resource)
		assert.Equal(t, "course", resource.Id.ResourceType)
	})

	t.Run("should use ID as display name when name is empty", func(t *testing.T) {
		course := client.Course{
			Id:          "course1",
			CourseTitle: "",
		}

		resource, err := courseResource(course, nil)

		require.NoError(t, err)
		assert.Equal(t, "course1", resource.DisplayName)
	})
}
