package client

import (
	"context"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type StatusesStore map[string]map[string]string

// Load given a Report (which again, can be on the order of 1 GB), and just
// create a mapping of course IDs to a mapping of user IDs to statuses. e.g.:
//
//	{
//	  "00000000-0000-0000-0000-000000000000": {
//	    "00000000-0000-0000-0000-000000000001": "in_progress",
//	    "00000000-0000-0000-0000-000000000002": "completed",
//	  },
//	}
func (r StatusesStore) Load(ctx context.Context, report *Report) error {
	logger := ctxzap.Extract(ctx)
	startTime := time.Now()

	logger.Debug("Starting to load status store from report",
		zap.Int("report_entries", len(*report)))

	totalEntries := 0
	uniqueCourses := 0
	uniqueUsers := make(map[string]bool)
	statusCounts := make(map[string]int)

	for _, row := range *report {
		found, ok := r[row.ContentId]
		if !ok {
			found = make(map[string]string)
			uniqueCourses++
		}

		status := toStatus(row.Status)
		found[row.UserId] = status
		r[row.ContentId] = found

		uniqueUsers[row.UserId] = true
		statusCounts[status]++
		totalEntries++
	}

	logger.Info("Status store loaded successfully",
		zap.Int("total_entries", totalEntries),
		zap.Int("unique_courses", uniqueCourses),
		zap.Int("unique_users", len(uniqueUsers)),
		zap.Any("status_distribution", statusCounts),
		zap.Duration("duration", time.Since(startTime)))

	return nil
}

// Get - return a mapping of user IDs to course completion status.
// TODO(marcos) Should we use enums instead?
func (r StatusesStore) Get(courseUUID string) map[string]string {
	found, ok := r[courseUUID]
	if !ok {
		// `nil` and empty map are equivalent.
		return nil
	}
	return found
}

func toStatus(status string) string {
	switch status {
	case "":
		return "no_status_reported"
	case "Completed", "Achieved", "Listened", "Read", "Watched":
		return "completed"
	case "Started", "Active":
		return "in_progress"
	default:
		return "status_undefined"
	}
}
