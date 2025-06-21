package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/iiiatthew/baton-percipio-report/pkg/config"
	"go.uber.org/zap"
)

const (
	ApiPathLearningActivityReport = "/reporting/v1/organizations/%s/report-requests/learning-activity"
	ApiPathReport                 = "/reporting/v1/organizations/%s/report-requests/%s"
	BaseApiUrl                    = "https://api.percipio.com"
)

type Client struct {
	baseUrl        *url.URL
	bearerToken    string
	StatusesStore  StatusesStore
	organizationId string
	ReportStatus   ReportStatus
	wrapper        *uhttp.BaseHttpClient
	loadedReport   *Report // Store the loaded report data
}

func New(
	ctx context.Context,
	baseUrl string,
	organizationId string,
	token string,
) (*Client, error) {
	httpClient, err := uhttp.NewClient(
		ctx,
		uhttp.WithLogger(
			true,
			ctxzap.Extract(ctx),
		),
	)
	if err != nil {
		return nil, err
	}

	parsedUrl, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}

	wrapper, err := uhttp.NewBaseHttpClientWithContext(ctx, httpClient)
	if err != nil {
		return nil, err
	}

	return &Client{
		StatusesStore:  make(map[string]map[string]string),
		baseUrl:        parsedUrl,
		bearerToken:    token,
		organizationId: organizationId,
		wrapper:        wrapper,
	}, nil
}

// GenerateLearningActivityReport makes a post request to the API asking it to
// start generating a report. We'll need to then poll a different endpoint to
// get the actual report data.
func (c *Client) GenerateLearningActivityReport(
	ctx context.Context,
	lookbackPeriod time.Duration,
) (
	*v2.RateLimitDescription,
	error,
) {
	logger := ctxzap.Extract(ctx)
	now := time.Now()

	reportStart := now.Add(-lookbackPeriod)

	body := ReportConfigurations{
		End:         now,
		Start:       reportStart,
		ContentType: "Course,Assessment",
	}

	logger.Info("Initiating learning activity report generation",
		zap.Time("report_start_date", reportStart),
		zap.Time("report_end_date", now),
		zap.Duration("lookback_period", lookbackPeriod),
		zap.String("content_types", body.ContentType))

	var target ReportStatus
	response, ratelimitData, err := c.post(
		ctx,
		ApiPathLearningActivityReport,
		body,
		&target,
	)
	if err != nil {
		logger.Error("Failed to initiate report generation", zap.Error(err))
		return ratelimitData, err
	}
	defer response.Body.Close()

	// Should include ID and "PENDING".
	c.ReportStatus = target

	logger.Debug("Report generation initiated",
		zap.String("report_id", target.Id),
		zap.String("report_status", target.Status))

	return ratelimitData, nil
}

// pollReportStatus uses standard net/http to poll report status without any caching
// Returns the report data directly when ready, avoiding a second HTTP call
func (c *Client) pollReportStatus(ctx context.Context) (*Report, error) {
	logger := ctxzap.Extract(ctx)

	// Build the URL for status polling
	statusURL := fmt.Sprintf("%s%s",
		c.baseUrl.String(),
		fmt.Sprintf(ApiPathReport, c.organizationId, c.ReportStatus.Id))

	// Create a simple HTTP client with no caching
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	var attempts int
	for i := range config.RetryAttemptsMaximum {
		attempts = i + 1

		req, err := http.NewRequestWithContext(ctx, "GET", statusURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create status request: %w", err)
		}

		// Add authorization header
		req.Header.Set("Authorization", "Bearer "+c.bearerToken)
		req.Header.Set("Accept", "application/json")

		logger.Debug("Polling report status (no cache)",
			zap.String("url", statusURL),
			zap.Int("attempt", attempts),
			zap.String("report_id", c.ReportStatus.Id))

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to poll report status: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read status response: %w", err)
		}

		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("status polling failed with code %d: %s", resp.StatusCode, string(body))
		}

		// Try to unmarshal as status first
		var status ReportStatus
		err = json.Unmarshal(body, &status)
		if err == nil {
			// Preserve the original report ID if the status response doesn't include it
			originalId := c.ReportStatus.Id
			c.ReportStatus = status
			if c.ReportStatus.Id == "" {
				c.ReportStatus.Id = originalId
			}

			logger.Debug("Report status update",
				zap.String("status", status.Status),
				zap.Int("attempt", attempts),
				zap.String("report_id", c.ReportStatus.Id))

			if status.Status == "FAILED" {
				return nil, fmt.Errorf("report generation failed: %v", status)
			}

			if status.Status == "COMPLETED" {
				logger.Info("Report generation completed",
					zap.String("report_id", c.ReportStatus.Id),
					zap.Int("polling_attempts", attempts))
				return nil, nil // Status completed but we need to fetch data separately
			}

			// Still processing, wait and continue
			time.Sleep(config.RetryAfterSeconds * time.Second)
			continue
		}

		// If we can't unmarshal as status, try as report data (report is ready)
		var report Report
		err = json.Unmarshal(body, &report)
		if err == nil {
			logger.Info("Report data ready immediately",
				zap.String("report_id", c.ReportStatus.Id),
				zap.Int("polling_attempts", attempts),
				zap.Int("report_entries", len(report)))
			return &report, nil
		}

		// Neither status nor report format worked
		sampleLen := len(body)
		if sampleLen > 100 {
			sampleLen = 100
		}
		logger.Debug("Response format not recognized, continuing",
			zap.String("response_sample", string(body[:sampleLen])))
		time.Sleep(config.RetryAfterSeconds * time.Second)
	}

	return nil, fmt.Errorf("report polling timed out after %d attempts", attempts)
}

func (c *Client) GetLearningActivityReport(
	ctx context.Context,
) (
	*v2.RateLimitDescription,
	error,
) {
	logger := ctxzap.Extract(ctx)

	// Poll for status using standard HTTP (no cache)
	report, err := c.pollReportStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to poll report status: %w", err)
	}

	var ratelimitData *v2.RateLimitDescription

	// If polling returned report data directly, use it
	if report != nil {
		logger.Debug("Using report data from polling response")
		c.loadedReport = report
		c.ReportStatus.Status = "done"
		// Create a basic rate limit response since polling doesn't provide it
		ratelimitData = &v2.RateLimitDescription{
			Limit:     -1,
			Remaining: 0,
			ResetAt:   nil,
			Status:    v2.RateLimitDescription_STATUS_UNSPECIFIED,
		}
	} else {
		// Otherwise fetch the report data
		var target Report
		response, rateLimit, err := c.get(
			ctx,
			fmt.Sprintf(ApiPathReport, "%s", c.ReportStatus.Id),
			nil,
			&target,
		)
		if err != nil {
			return rateLimit, fmt.Errorf("failed to fetch report data: %w", err)
		}
		defer response.Body.Close()

		logger.Debug("Fetched report data using baton-sdk")
		c.loadedReport = &target
		c.ReportStatus.Status = "done"
		ratelimitData = rateLimit
	}

	logger.Info("Report ready, loading data",
		zap.Int("report_entries", len(*c.loadedReport)))

	// Calculate approximate size
	reportSizeBytes := 0
	for _, entry := range *c.loadedReport {
		// Rough estimation of bytes per entry
		reportSizeBytes += len(entry.UserId) + len(entry.EmailAddress) + len(entry.FirstName) +
			len(entry.LastName) + len(entry.ContentId) + len(entry.ContentTitle) +
			len(entry.Status) + 100 // overhead for other fields
	}

	logger.Debug("Report data statistics",
		zap.Int("entries", len(*c.loadedReport)),
		zap.Int("estimated_size_bytes", reportSizeBytes),
		zap.Float64("estimated_size_mb", float64(reportSizeBytes)/1024/1024))

	err = c.StatusesStore.Load(ctx, c.loadedReport)
	if err != nil {
		return ratelimitData, err
	}
	return ratelimitData, nil
}

// GetLoadedReport returns the loaded report data.
func (c *Client) GetLoadedReport() *Report {
	return c.loadedReport
}
