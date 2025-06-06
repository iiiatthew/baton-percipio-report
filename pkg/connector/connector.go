package connector

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/iiiatthew/baton-percipio-report/pkg/client"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type Connector struct {
	client             *client.Client
	report             *client.Report
	reportLookback     time.Duration
	reportInitialized  bool
}

// ResourceSyncers returns a ResourceSyncer for each resource type that should be synced from the upstream service.
func (d *Connector) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncer {
	_ = ctx // This method returns static resource syncers
	return []connectorbuilder.ResourceSyncer{
		newUserBuilder(d.client, d.report, d),
		newCourseBuilder(d.client, d.report, d),
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
	_, err := d.client.GenerateLearningActivityReport(ctx, d.reportLookback)
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
	reportLookback time.Duration,
) (*Connector, error) {
	logger := ctxzap.Extract(ctx)
	logger.Info("Initializing Percipio connector",
		zap.String("organizationId", organizationID))

	percipioClient, err := client.New(
		ctx,
		client.BaseApiUrl,
		organizationID,
		token,
	)
	if err != nil {
		logger.Error("Failed to create Percipio client", zap.Error(err))
		return nil, err
	}

	connector := &Connector{
		client:         percipioClient,
		reportLookback: reportLookback,
	}

	return connector, nil
}

// ensureReportInitialized generates and loads the report if not already done for this sync
func (d *Connector) ensureReportInitialized(ctx context.Context) error {
	if d.reportInitialized {
		return nil
	}

	logger := ctxzap.Extract(ctx)
	logger.Info("Starting learning activity report generation for sync")
	reportGenStart := time.Now()
	
	_, err := d.client.GenerateLearningActivityReport(ctx, d.reportLookback)
	if err != nil {
		logger.Error("Failed to generate learning activity report",
			zap.Error(err),
			zap.Duration("duration", time.Since(reportGenStart)))
		return fmt.Errorf("failed to generate learning activity report: %w", err)
	}

	logger.Debug("Report generation request submitted",
		zap.Duration("duration", time.Since(reportGenStart)))

	reportLoadStart := time.Now()
	_, err = d.client.GetLearningActivityReport(ctx)
	if err != nil {
		logger.Error("Failed to retrieve learning activity report",
			zap.Error(err),
			zap.Duration("generation_duration", time.Since(reportGenStart)),
			zap.Duration("load_duration", time.Since(reportLoadStart)))
		return fmt.Errorf("failed to retrieve learning activity report: %w", err)
	}

	// Store the loaded report data
	d.report = d.client.GetLoadedReport()
	
	reportSize := 0
	if d.report != nil {
		reportSize = len(*d.report)
	}
	
	logger.Info("Learning activity report loaded successfully for sync",
		zap.Int("report_entries", reportSize),
		zap.Duration("total_duration", time.Since(reportGenStart)),
		zap.Duration("load_duration", time.Since(reportLoadStart)))

	d.reportInitialized = true
	return nil
}
