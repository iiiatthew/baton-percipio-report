package client

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
func (r StatusesStore) Load(report *Report) error {
	for _, row := range *report {
		// Use ContentId as the key instead of ContentUUID
		found, ok := r[row.ContentId]
		if !ok {
			found = make(map[string]string)
		}

		found[row.UserId] = toStatus(row.Status)
		r[row.ContentId] = found
	}

	return nil
}

// Get - return a mapping of user IDs to course completion status.
// TODO(marcos) Should we use enums instead?
func (r StatusesStore) Get(courseId string) map[string]string {
	found, ok := r[courseId]
	if !ok {
		// `nil` and empty map are equivalent.
		return nil
	}
	return found
}

// toStatus convert Percipio status to C1 status.
func toStatus(status string) string {
	switch status {
	case "Completed", "Achieved", "Listened", "Read", "Watched":
		return "completed"
	case "Started", "Active":
		return "in_progress"
	default:
		return "in_progress" // Default to in_progress for any other status
	}
}
