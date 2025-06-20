package connector

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/iiiatthew/baton-percipio-report/pkg/client"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

// ReportState represents the current state of report generation
type ReportState int

const (
	ReportNotStarted ReportState = iota
	ReportInProgress
	ReportCompleted
	ReportFailed
)

type Connector struct {
	client         *client.Client
	report         *client.Report
	reportLookback time.Duration
	reportState    ReportState
	reportMutex    sync.RWMutex
	reportError    error
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
	// Always generate the report during validation - this happens once per sync cycle
	err := d.generateReport(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to validate Percipio credentials: %w", err)
	}
	return nil, nil
}

// waitForReport waits for the report to be available or returns an error if generation failed
func (d *Connector) waitForReport(ctx context.Context) error {
	d.reportMutex.RLock()
	defer d.reportMutex.RUnlock()

	switch d.reportState {
	case ReportCompleted:
		return nil
	case ReportFailed:
		return d.reportError
	case ReportNotStarted, ReportInProgress:
		return fmt.Errorf("report not ready: current state is %d", d.reportState)
	default:
		return fmt.Errorf("unknown report state: %d", d.reportState)
	}
}

// generateReport creates and loads the learning activity report
func (d *Connector) generateReport(ctx context.Context) error {
	d.reportMutex.Lock()
	defer d.reportMutex.Unlock()

	// If already completed, return success
	if d.reportState == ReportCompleted {
		return nil
	}

	// If currently in progress or failed, reset and try again
	d.reportState = ReportInProgress
	d.reportError = nil

	logger := ctxzap.Extract(ctx)
	logger.Info("Starting learning activity report generation for sync")
	reportGenStart := time.Now()

	_, err := d.client.GenerateLearningActivityReport(ctx, d.reportLookback)
	if err != nil {
		d.reportState = ReportFailed
		d.reportError = fmt.Errorf("failed to generate learning activity report: %w", err)
		logger.Error("Failed to generate learning activity report",
			zap.Error(err),
			zap.Duration("duration", time.Since(reportGenStart)))
		return d.reportError
	}

	logger.Debug("Report generation request submitted",
		zap.Duration("duration", time.Since(reportGenStart)))

	reportLoadStart := time.Now()
	_, err = d.client.GetLearningActivityReport(ctx)
	if err != nil {
		d.reportState = ReportFailed
		d.reportError = fmt.Errorf("failed to retrieve learning activity report: %w", err)
		logger.Error("Failed to retrieve learning activity report",
			zap.Error(err),
			zap.Duration("generation_duration", time.Since(reportGenStart)),
			zap.Duration("load_duration", time.Since(reportLoadStart)))
		return d.reportError
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

	d.reportState = ReportCompleted
	return nil
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
