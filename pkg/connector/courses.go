package connector

import (
	"context"
	"fmt"

	"github.com/iiiatthew/baton-percipio-report/pkg/client"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	resourceSdk "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

const (
	assignedEntitlement   = "assigned"
	completedEntitlement  = "completed"
	inProgressEntitlement = "in_progress"
)

type courseBuilder struct {
	client       *client.Client
	resourceType *v2.ResourceType
	report       *client.Report
	connector    *Connector
}

func (o *courseBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	_ = ctx // This method returns a static resource type
	return o.resourceType
}

// Create a new connector resource for a Percipio course.
func courseResource(course client.Course, parentResourceID *v2.ResourceId) (*v2.Resource, error) {
	resourceOpts := []resourceSdk.ResourceOption{
		resourceSdk.WithParentResourceID(parentResourceID),
	}

	// Use course name, fallback to ID if name is empty
	displayName := course.CourseTitle
	if displayName == "" {
		displayName = course.CourseTitle
	}

	resource, err := resourceSdk.NewResource(
		displayName,
		courseResourceType,
		course.Id,
		resourceOpts...,
	)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

func (o *courseBuilder) List(
	ctx context.Context,
	parentResourceID *v2.ResourceId,
	_ *pagination.Token,
) (
	[]*v2.Resource,
	string,
	annotations.Annotations,
	error,
) {
	logger := ctxzap.Extract(ctx)
	logger.Debug("Starting Courses List from Report Data")

	outputResources := make([]*v2.Resource, 0)
	var outputAnnotations annotations.Annotations

	// Ensure report is initialized for this sync
	if err := o.connector.ensureReportInitialized(ctx); err != nil {
		return nil, "", outputAnnotations, err
	}

	if o.connector.report == nil || len(*o.connector.report) == 0 {
		logger.Warn("No report data available")
		return outputResources, "", outputAnnotations, nil
	}

	// Extract unique courses from report
	courseMap := make(map[string]client.Course)
	for _, entry := range *o.connector.report {
		courseId := entry.ContentUUID
		if courseId == "" {
			logger.Warn("Course missing contentUUID, skipping",
				zap.String("contentTitle", entry.ContentTitle),
				zap.String("contentUuid", entry.ContentUUID),
				zap.String("contentType", entry.ContentType))
			continue
		}

		if _, exists := courseMap[courseId]; !exists {
			// Check for missing title
			if entry.ContentTitle == "" {
				logger.Warn("Course missing title",
					zap.String("courseId", courseId),
					zap.String("contentType", entry.ContentType))
			}

			courseMap[courseId] = client.Course{
				Id:          courseId,
				CourseTitle: entry.ContentTitle,
			}
		}
	}

	// Convert map to resources
	for _, course := range courseMap {
		courseResource0, err := courseResource(course, parentResourceID)
		if err != nil {
			return nil, "", outputAnnotations, err
		}
		outputResources = append(outputResources, courseResource0)
	}

	// Log deduplication statistics
	totalDuplicates := len(*o.connector.report) - len(courseMap)
	logger.Info("Course extraction completed",
		zap.Int("total_report_entries", len(*o.connector.report)),
		zap.Int("unique_courses", len(outputResources)),
		zap.Int("duplicate_entries", totalDuplicates),
		zap.Float64("deduplication_ratio", float64(totalDuplicates)/float64(len(*o.connector.report))))

	// No pagination needed since we're returning all courses from the report
	return outputResources, "", outputAnnotations, nil
}

func (o *courseBuilder) Entitlements(
	_ context.Context,
	resource *v2.Resource,
	_ *pagination.Token,
) (
	[]*v2.Entitlement,
	string,
	annotations.Annotations,
	error,
) {
	return []*v2.Entitlement{
		entitlement.NewAssignmentEntitlement(
			resource,
			assignedEntitlement,
			entitlement.WithGrantableTo(userResourceType),
			entitlement.WithDisplayName(fmt.Sprintf("Course %s %s", resource.DisplayName, assignedEntitlement)),
			entitlement.WithDescription(fmt.Sprintf("Assigned course %s in Percipio", resource.DisplayName)),
		),
		entitlement.NewAssignmentEntitlement(
			resource,
			completedEntitlement,
			entitlement.WithGrantableTo(userResourceType),
			entitlement.WithDisplayName(fmt.Sprintf("Course %s %s", resource.DisplayName, completedEntitlement)),
			entitlement.WithDescription(fmt.Sprintf("Completed course %s in Percipio", resource.DisplayName)),
		),
		entitlement.NewAssignmentEntitlement(
			resource,
			inProgressEntitlement,
			entitlement.WithGrantableTo(userResourceType),
			entitlement.WithDisplayName(fmt.Sprintf("Course %s %s", resource.DisplayName, inProgressEntitlement)),
			entitlement.WithDescription(fmt.Sprintf("In progress course %s in Percipio", resource.DisplayName)),
		),
	}, "", nil, nil
}

// Grants returns the grants for a course resource based on the pre-loaded report data.
func (o *courseBuilder) Grants(
	ctx context.Context,
	resource *v2.Resource,
	_ *pagination.Token,
) (
	[]*v2.Grant,
	string,
	annotations.Annotations,
	error,
) {
	logger := ctxzap.Extract(ctx)
	var outputAnnotations annotations.Annotations

	// Report is already loaded during sync initialization
	// Just get the status map for this course
	statusesMap := o.client.StatusesStore.Get(resource.Id.Resource)

	logger.Debug("Looking up grants for course",
		zap.String("course_id", resource.Id.Resource),
		zap.String("course_name", resource.DisplayName),
		zap.Int("grant_count", len(statusesMap)))

	grants := make([]*v2.Grant, 0)
	statusCounts := make(map[string]int)

	for userId, status := range statusesMap {
		principalId, err := resourceSdk.NewResourceID(userResourceType, userId)
		if err != nil {
			logger.Error("Failed to create principal ID",
				zap.Error(err),
				zap.String("user_id", userId),
				zap.String("course_id", resource.Id.Resource))
			return nil, "", outputAnnotations, err
		}
		nextGrant := grant.NewGrant(resource, status, principalId)
		grants = append(grants, nextGrant)
		statusCounts[status]++
	}

	if len(grants) > 0 {
		logger.Debug("Grants created for course",
			zap.String("course_id", resource.Id.Resource),
			zap.Int("total_grants", len(grants)),
			zap.Any("status_distribution", statusCounts))
	}

	return grants, "", outputAnnotations, nil
}

func newCourseBuilder(client *client.Client, report *client.Report, connector *Connector) *courseBuilder {
	return &courseBuilder{
		client:       client,
		resourceType: courseResourceType,
		report:       report,
		connector:    connector,
	}
}
