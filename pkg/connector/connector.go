package connector

import (
	"context"
	"fmt"
	"io"

	"github.com/iiiatthew/baton-percipio-report/pkg/client"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
)

type Connector struct {
	client *client.Client
	report *client.Report
}

// ResourceSyncers returns a ResourceSyncer for each resource type that should be synced from the upstream service.
func (d *Connector) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncer {
	_ = ctx // This method returns static resource syncers
	return []connectorbuilder.ResourceSyncer{
		newUserBuilder(d.client, d.report),
		newCourseBuilder(d.client, d.report),
	}
}

// Asset takes an input AssetRef and attempts to fetch it using the connector's authenticated http client
// It streams a response, always starting with a metadata object, following by chunked payloads for the asset.
func (d *Connector) Asset(ctx context.Context, asset *v2.AssetRef) (string, io.ReadCloser, error) {
	_ = ctx // Asset streaming not implemented for this connector
	_ = asset
	return "", nil, nil
}

// Metadata returns metadata about the connector.
func (d *Connector) Metadata(ctx context.Context) (*v2.ConnectorMetadata, error) {
	_ = ctx // This method returns static metadata
	return &v2.ConnectorMetadata{
		DisplayName: "Percipio Connector",
		Description: "Connector syncing users from Percipio",
	}, nil
}

// Validate is called to ensure that the connector is properly configured. It should exercise any API credentials
// to be sure that they are valid.
func (d *Connector) Validate(ctx context.Context) (annotations.Annotations, error) {
	// Test the connection by attempting to generate a report status check
	// This validates that our credentials and organization ID are correct
	_, err := d.client.GenerateLearningActivityReport(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to validate Percipio credentials: %w", err)
	}
	return nil, nil
}

// New returns a new instance of the connector.
func New(
	ctx context.Context,
	organizationID string,
	token string,
) (*Connector, error) {
	percipioClient, err := client.New(
		ctx,
		client.BaseApiUrl,
		organizationID,
		token,
	)
	if err != nil {
		return nil, err
	}

	connector := &Connector{
		client: percipioClient,
	}

	// Generate and load the report during initialization
	_, err = percipioClient.GenerateLearningActivityReport(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate learning activity report: %w", err)
	}

	_, err = percipioClient.GetLearningActivityReport(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve learning activity report: %w", err)
	}

	// Store the loaded report data
	connector.report = percipioClient.GetLoadedReport()

	return connector, nil
}
