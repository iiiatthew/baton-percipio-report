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
			reportState: ReportCompleted,
			report: &client.Report{
				{
					UserId:       "michael.bolton@initech.com",
					ContentId:    "bs_adg02_a23_enus",
					ContentTitle: "Case Studies: Successful Data Privacy Implementations",
					ContentType:  "Course",
					Status:       "Completed",
				},
				{
					UserId:       "milton.waddams@initech.com",
					ContentId:    "bs_adg02_a23_enus",
					ContentTitle: "Case Studies: Successful Data Privacy Implementations",
					ContentType:  "Course",
					Status:       "Started",
				},
				{
					UserId:       "michael.bolton@initech.com",
					ContentId:    "another_course_id",
					ContentTitle: "Advanced Go Patterns",
					ContentType:  "Assessment",
					Status:       "Active",
				},
				{
					UserId:       "peter.gibbons@initech.com",
					ContentId:    "bs_adg02_a23_enus",
					ContentTitle: "Case Studies: Successful Data Privacy Implementations",
					ContentType:  "Course",
					Status:       "",
				},
				{
					UserId:       "bill.lumbergh@initech.com",
					ContentId:    "another_course_id",
					ContentTitle: "Advanced Go Patterns",
					ContentType:  "Assessment",
					Status:       "UnknownStatus",
				},
			},
		}

		c := newCourseBuilder(nil, nil, connector)

		resources, nextToken, annotations, err := c.List(ctx, nil, &pagination.Token{})

		require.NoError(t, err)
		assert.Empty(t, nextToken)
		test.AssertNoRatelimitAnnotations(t, annotations)
		require.Len(t, resources, 2)

		course1 := findResourceById(resources, "bs_adg02_a23_enus")
		require.NotNil(t, course1)
		assert.Equal(t, "Case Studies: Successful Data Privacy Implementations (Course)", course1.DisplayName)

		course2 := findResourceById(resources, "another_course_id")
		require.NotNil(t, course2)
		assert.Equal(t, "Advanced Go Patterns (Assessment)", course2.DisplayName)
	})

	t.Run("should handle missing contentId", func(t *testing.T) {
		connector := &Connector{
			reportState: ReportCompleted,
			report: &client.Report{
				{
					UserId:       "michael.bolton@initech.com",
					ContentId:    "",
					ContentTitle: "Course Without ID",
					ContentType:  "Course",
					Status:       "Completed",
				},
				{
					UserId:       "michael.bolton@initech.com",
					ContentId:    "bs_adg02_a23_enus",
					ContentTitle: "Case Studies: Successful Data Privacy Implementations",
					ContentType:  "Course",
					Status:       "Completed",
				},
			},
		}

		c := newCourseBuilder(nil, nil, connector)

		resources, _, _, err := c.List(ctx, nil, &pagination.Token{})

		require.NoError(t, err)
		require.Len(t, resources, 1)
		assert.Equal(t, "bs_adg02_a23_enus", resources[0].Id.Resource)
	})

	t.Run("should handle missing title", func(t *testing.T) {
		connector := &Connector{
			reportState: ReportCompleted,
			report: &client.Report{
				{
					UserId:       "michael.bolton@initech.com",
					ContentId:    "bs_adg02_a23_enus",
					ContentTitle: "",
					ContentType:  "Course",
					Status:       "Completed",
				},
			},
		}

		c := newCourseBuilder(nil, nil, connector)

		resources, _, _, err := c.List(ctx, nil, &pagination.Token{})

		require.NoError(t, err)
		require.Len(t, resources, 1)
		assert.Equal(t, " (Course)", resources[0].DisplayName)
	})

	t.Run("should wait for report generation", func(t *testing.T) {
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
					ContentId:    "test-course",
					ContentTitle: "Test Course",
					ContentType:  "Course",
					Status:       "Completed",
				},
			},
		}

		c := newCourseBuilder(percipioClient, nil, connector)

		resources, _, _, err := c.List(ctx, nil, &pagination.Token{})

		require.NoError(t, err)
		assert.Equal(t, ReportCompleted, connector.reportState)
		assert.Greater(t, len(resources), 0)
	})
}

func TestCoursesEntitlements(t *testing.T) {
	ctx := context.Background()

	c := newCourseBuilder(nil, nil, nil)
	course := &v2.Resource{
		DisplayName: "Case Studies: Successful Data Privacy Implementations (Course)",
		Id: &v2.ResourceId{
			ResourceType: "course",
			Resource:     "bs_adg02_a23_enus",
		},
	}

	entitlements, nextToken, annotations, err := c.Entitlements(ctx, course, &pagination.Token{})

	require.NoError(t, err)
	assert.Empty(t, nextToken)
	assert.Nil(t, annotations)
	require.Len(t, entitlements, 5)

	entitlementSlugs := make([]string, len(entitlements))
	for i, ent := range entitlements {
		entitlementSlugs[i] = ent.Slug
	}
	assert.Contains(t, entitlementSlugs, "assigned")
	assert.Contains(t, entitlementSlugs, "completed")
	assert.Contains(t, entitlementSlugs, "in_progress")
	assert.Contains(t, entitlementSlugs, "no_status_reported")
	assert.Contains(t, entitlementSlugs, "status_undefined")
}

func TestCoursesGrants(t *testing.T) {
	ctx := context.Background()

	t.Run("should return grants for course", func(t *testing.T) {
		statusStore := make(client.StatusesStore)
		statusStore["bs_adg02_a23_enus"] = map[string]string{
			"michael.bolton@initech.com": "completed",
			"milton.waddams@initech.com": "in_progress",
			"peter.gibbons@initech.com":  "no_status_reported",
			"bill.lumbergh@initech.com":  "status_undefined",
		}

		percipioClient := &client.Client{
			StatusesStore: statusStore,
		}

		c := newCourseBuilder(percipioClient, nil, nil)
		course := &v2.Resource{
			DisplayName: "Case Studies: Successful Data Privacy Implementations (Course)",
			Id: &v2.ResourceId{
				ResourceType: "course",
				Resource:     "bs_adg02_a23_enus",
			},
		}

		grants, nextToken, annotations, err := c.Grants(ctx, course, &pagination.Token{})

		require.NoError(t, err)
		assert.Empty(t, nextToken)
		test.AssertNoRatelimitAnnotations(t, annotations)
		require.Len(t, grants, 4)

		grantsByUser := make(map[string]string)
		for _, grant := range grants {
			entitlementId := grant.Entitlement.Id
			parts := strings.Split(entitlementId, ":")
			if len(parts) >= 3 {
				entitlementName := parts[len(parts)-1]
				grantsByUser[grant.Principal.Id.Resource] = entitlementName
			}
		}
		assert.Equal(t, "completed", grantsByUser["michael.bolton@initech.com"])
		assert.Equal(t, "in_progress", grantsByUser["milton.waddams@initech.com"])
		assert.Equal(t, "no_status_reported", grantsByUser["peter.gibbons@initech.com"])
		assert.Equal(t, "status_undefined", grantsByUser["bill.lumbergh@initech.com"])
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
			Id:          "bs_adg02_a23_enus",
			CourseTitle: "Case Studies: Successful Data Privacy Implementations (Course)",
		}

		resource, err := courseResource(course, nil)

		require.NoError(t, err)
		assert.Equal(t, "Case Studies: Successful Data Privacy Implementations (Course)", resource.DisplayName)
		assert.Equal(t, "bs_adg02_a23_enus", resource.Id.Resource)
		assert.Equal(t, "course", resource.Id.ResourceType)
	})

	t.Run("should use title as display name when name is empty", func(t *testing.T) {
		course := client.Course{
			Id:          "bs_adg02_a23_enus",
			CourseTitle: "",
		}

		resource, err := courseResource(course, nil)

		require.NoError(t, err)
		assert.Equal(t, "", resource.DisplayName)
	})
}
