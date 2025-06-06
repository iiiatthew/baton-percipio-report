package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
// start generating a report. We'll need to then poll a _different_ endpoint to
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

func (c *Client) GetLearningActivityReport(
	ctx context.Context,
) (
	*v2.RateLimitDescription,
	error,
) {
	var (
		ratelimitData *v2.RateLimitDescription
		target        Report
	)

	l := ctxzap.Extract(ctx)
	var attempts int
	for i := 0; i < config.RetryAttemptsMaximum; i++ {
		attempts = i + 1
		// Use an anonymous function to ensure proper resource cleanup with defer
		shouldBreak := false
		err := func() error {
			// While the report is still processing, we get this ReportStatus
			// object. Once we actually get data, it'll return an array of rows.
			response, ratelimitData0, err := c.get(
				ctx,
				// Punt setting `organizationId`, it is added in `doRequest()`.
				fmt.Sprintf(ApiPathReport, "%s", c.ReportStatus.Id),
				nil,
				// Don't use response body because Percipio's API closes connections early and returns EOF sometimes.
				nil,
			)
			ratelimitData = ratelimitData0
			if err != nil {
				l.Error("error getting report", zap.Error(err))
				// Ignore unexpected EOF because Precipio returns this on success sometimes
				if !errors.Is(err, io.ErrUnexpectedEOF) {
					return err
				}
				// Continue to next iteration on unexpected EOF
				return nil
			}
			if response == nil {
				return fmt.Errorf("no response from precipio api")
			}

			defer response.Body.Close()
			bodyBytes, err := io.ReadAll(response.Body)
			if err != nil {
				l.Error("error reading response body", zap.Error(err))
				return err
			}

			// Response can be a report status if the report isn't done processing, or the report. Try status first.
			err = json.Unmarshal(bodyBytes, &c.ReportStatus)
			if err == nil {
				l.Debug("Report still processing, will retry",
					zap.String("status", c.ReportStatus.Status),
					zap.Int("attempt", i+1),
					zap.Int("max_attempts", config.RetryAttemptsMaximum),
					zap.Duration("wait_before_retry", config.RetryAfterSeconds*time.Second),
					zap.String("report_id", c.ReportStatus.Id))
				if c.ReportStatus.Status == "FAILED" {
					return fmt.Errorf("report generation failed: %v", c.ReportStatus)
				}
				// Report is still processing, continue to next iteration
				return nil
			}
			syntaxError := new(json.SyntaxError)
			if errors.As(err, &syntaxError) {
				l.Warn("syntax error unmarshaling report status", zap.Error(err))
				// Continue to next iteration
				return nil
			}
			unmarshalError := new(json.UnmarshalTypeError)
			if !errors.As(err, &unmarshalError) {
				return err
			}

			l.Debug("unmarshaling to report status failed. trying to unmarshall as report", zap.Error(err))
			err = json.Unmarshal(bodyBytes, &target)
			if err != nil {
				return err
			}
			// We got the report object.
			shouldBreak = true
			return nil
		}()

		if err != nil {
			return ratelimitData, err
		}

		if shouldBreak {
			break
		}

		time.Sleep(config.RetryAfterSeconds * time.Second)
	}

	c.ReportStatus.Status = "done"
	l.Info("Report ready, loading data",
		zap.Int("report_entries", len(target)),
		zap.Int("polling_attempts", attempts),
		zap.Duration("wait_time", time.Duration(attempts-1)*config.RetryAfterSeconds*time.Second))

	// Store the raw report data
	c.loadedReport = &target
	
	// Calculate approximate size
	reportSizeBytes := 0
	for _, entry := range target {
		// Rough estimation of bytes per entry
		reportSizeBytes += len(entry.UserId) + len(entry.EmailAddress) + 
			len(entry.FirstName) + len(entry.LastName) +
			len(entry.ContentId) + len(entry.ContentTitle) +
			len(entry.Status) + 100 // overhead for other fields
	}
	
	l.Debug("Report data statistics",
		zap.Int("entries", len(target)),
		zap.Int("estimated_size_bytes", reportSizeBytes),
		zap.Float64("estimated_size_mb", float64(reportSizeBytes)/1024/1024))

	err := c.StatusesStore.Load(ctx, &target)
	if err != nil {
		return ratelimitData, err
	}
	return ratelimitData, nil
}

// GetLoadedReport returns the loaded report data.
func (c *Client) GetLoadedReport() *Report {
	return c.loadedReport
}
