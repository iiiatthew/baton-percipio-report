package connector

import (
	"context"
	"fmt"

	"github.com/iiiatthew/baton-percipio-report/pkg/client"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	resourceSdk "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type userBuilder struct {
	client       *client.Client
	resourceType *v2.ResourceType
	report       *client.Report
	connector    *Connector
}

func (o *userBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	_ = ctx // This method returns a static resource type
	return o.resourceType
}

func getDisplayName(user client.User) string {
	return fmt.Sprintf("%s %s", user.FirstName, user.LastName)
}

// Create a new connector resource for a Percipio user.
func userResource(user client.User, parentResourceID *v2.ResourceId) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"id":           user.Id,
		"login_id":     user.Id,
		"display_name": getDisplayName(user),
		"email":        user.Email,
		"first_name":   user.FirstName,
		"last_name":    user.LastName,
	}

	userTraitOptions := []resourceSdk.UserTraitOption{
		resourceSdk.WithEmail(user.Email, true),
		resourceSdk.WithStatus(v2.UserTrait_Status_STATUS_ENABLED),
		resourceSdk.WithUserProfile(profile),
	}

	userResource0, err := resourceSdk.NewUserResource(
		getDisplayName(user),
		userResourceType,
		user.Id,
		userTraitOptions,
		resourceSdk.WithParentResourceID(parentResourceID),
	)
	if err != nil {
		return nil, err
	}

	return userResource0, nil
}

// List returns all the users from the learning activity report as resource objects.
// Users include a UserTrait because they are the 'shape' of a standard user.
func (o *userBuilder) List(
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
	logger.Debug("Starting Users List from Report Data")

	outputResources := make([]*v2.Resource, 0)
	var outputAnnotations annotations.Annotations

	// Wait for report to be generated during validation
	if err := o.connector.waitForReport(ctx); err != nil {
		return nil, "", outputAnnotations, err
	}

	if o.connector.report == nil || len(*o.connector.report) == 0 {
		logger.Warn("No report data available")
		return outputResources, "", outputAnnotations, nil
	}

	// Extract unique users from report, keeping the most recent data
	type userWithDate struct {
		user           client.User
		mostRecentDate string
	}

	userMap := make(map[string]userWithDate)

	for _, entry := range *o.connector.report {
		// Skip entries with empty userId
		if entry.UserId == "" {
			continue
		}

		var mostRecentDate string
		switch {
		case entry.CompletedDate != "":
			mostRecentDate = entry.CompletedDate
		case entry.LastAccess != "":
			mostRecentDate = entry.LastAccess
		default:
			mostRecentDate = entry.FirstAccess
		}

		if existing, exists := userMap[entry.UserId]; exists {
			if mostRecentDate > existing.mostRecentDate {
				logger.Debug("Updating user with more recent data",
					zap.String("userId", entry.UserId),
					zap.String("oldDate", existing.mostRecentDate),
					zap.String("newDate", mostRecentDate))

				userMap[entry.UserId] = userWithDate{
					user: client.User{
						Id:        entry.UserId,
						Email:     entry.EmailAddress,
						FirstName: entry.FirstName,
						LastName:  entry.LastName,
					},
					mostRecentDate: mostRecentDate,
				}
			}
		} else {
			userMap[entry.UserId] = userWithDate{
				user: client.User{
					Id:        entry.UserId,
					Email:     entry.EmailAddress,
					FirstName: entry.FirstName,
					LastName:  entry.LastName,
				},
				mostRecentDate: mostRecentDate,
			}
		}
	}

	// Convert map to resources
	for _, userData := range userMap {
		userResource0, err := userResource(userData.user, parentResourceID)
		if err != nil {
			return nil, "", outputAnnotations, err
		}
		outputResources = append(outputResources, userResource0)
	}

	// Log deduplication statistics
	totalDuplicates := len(*o.connector.report) - len(userMap)
	logger.Info("User extraction completed",
		zap.Int("total_report_entries", len(*o.connector.report)),
		zap.Int("unique_users", len(outputResources)),
		zap.Int("duplicate_entries", totalDuplicates),
		zap.Float64("deduplication_ratio", float64(totalDuplicates)/float64(len(*o.connector.report))))

	// No pagination needed since we're returning all users from the report
	return outputResources, "", outputAnnotations, nil
}

// Entitlements always returns an empty slice for users.
func (o *userBuilder) Entitlements(
	_ context.Context,
	_ *v2.Resource,
	_ *pagination.Token,
) (
	[]*v2.Entitlement,
	string,
	annotations.Annotations,
	error,
) {
	return nil, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *userBuilder) Grants(
	_ context.Context,
	_ *v2.Resource,
	_ *pagination.Token,
) (
	[]*v2.Grant,
	string,
	annotations.Annotations,
	error,
) {
	return nil, "", nil, nil
}

func newUserBuilder(client *client.Client, report *client.Report, connector *Connector) *userBuilder {
	return &userBuilder{
		client:       client,
		resourceType: userResourceType,
		report:       report,
		connector:    connector,
	}
}
