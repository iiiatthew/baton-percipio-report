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
	displayName := course.Name
	if displayName == "" {
		displayName = course.Id
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

	if o.report == nil || len(*o.report) == 0 {
		logger.Warn("No report data available")
		return outputResources, "", outputAnnotations, nil
	}

	// Extract unique courses from report
	courseMap := make(map[string]client.Course)
	for _, entry := range *o.report {
		// Use contentId as the primary identifier
		courseId := entry.ContentId
		if courseId == "" {
			logger.Warn("Course missing contentId, skipping",
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
				Id:   courseId,
				Name: entry.ContentTitle,
				// UUID: entry.ContentUuid, // Commented out as per requirements
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

	logger.Debug("Extracted courses from report", zap.Int("courseCount", len(outputResources)))

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
	_ = ctx // Report data is pre-loaded, no API calls needed
	var outputAnnotations annotations.Annotations

	// Report is already loaded during connector initialization
	// Just get the status map for this course
	statusesMap := o.client.StatusesStore.Get(resource.Id.Resource)

	grants := make([]*v2.Grant, 0)
	for userId, status := range statusesMap {
		principalId, err := resourceSdk.NewResourceID(userResourceType, userId)
		if err != nil {
			return nil, "", outputAnnotations, err
		}
		nextGrant := grant.NewGrant(resource, status, principalId)
		grants = append(grants, nextGrant)
	}

	return grants, "", outputAnnotations, nil
}

func newCourseBuilder(client *client.Client, report *client.Report) *courseBuilder {
	return &courseBuilder{
		client:       client,
		resourceType: courseResourceType,
		report:       report,
	}
}
