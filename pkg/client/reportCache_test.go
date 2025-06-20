package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusesStoreLoad(t *testing.T) {
	ctx := context.Background()

	t.Run("should load report data correctly", func(t *testing.T) {
		store := make(StatusesStore)
		report := &Report{
			{
				UserId:       "michael.bolton@initech.com",
				ContentId:    "bs_adg02_a23_enus",
				ContentTitle: "Case Studies: Successful Data Privacy Implementations",
				Status:       "Completed",
				EmailAddress: "michael.bolton@initech.com",
			},
			{
				UserId:       "milton.waddams@initech.com",
				ContentId:    "bs_adg02_a23_enus",
				ContentTitle: "Case Studies: Successful Data Privacy Implementations",
				Status:       "Started",
				EmailAddress: "milton.waddams@initech.com",
			},
			{
				UserId:       "michael.bolton@initech.com",
				ContentId:    "another_course_id",
				ContentTitle: "Another Course",
				Status:       "Active",
				EmailAddress: "michael.bolton@initech.com",
			},
			{
				UserId:       "peter.gibbons@initech.com",
				ContentId:    "bs_adg02_a23_enus",
				ContentTitle: "Case Studies: Successful Data Privacy Implementations",
				Status:       "",
				EmailAddress: "peter.gibbons@initech.com",
			},
			{
				UserId:       "bill.lumbergh@initech.com",
				ContentId:    "bs_adg02_a23_enus",
				ContentTitle: "Case Studies: Successful Data Privacy Implementations",
				Status:       "UnknownStatus",
				EmailAddress: "bill.lumbergh@initech.com",
			},
		}

		err := store.Load(ctx, report)
		require.NoError(t, err)

		assert.Len(t, store, 2)

		course1Users := store.Get("bs_adg02_a23_enus")
		assert.Len(t, course1Users, 4)
		assert.Equal(t, "completed", course1Users["michael.bolton@initech.com"])
		assert.Equal(t, "in_progress", course1Users["milton.waddams@initech.com"])
		assert.Equal(t, "no_status_reported", course1Users["peter.gibbons@initech.com"])
		assert.Equal(t, "status_undefined", course1Users["bill.lumbergh@initech.com"])

		course2Users := store.Get("another_course_id")
		assert.Len(t, course2Users, 1)
		assert.Equal(t, "in_progress", course2Users["michael.bolton@initech.com"])
	})

	t.Run("should handle empty report", func(t *testing.T) {
		store := make(StatusesStore)
		report := &Report{}

		err := store.Load(ctx, report)
		require.NoError(t, err)
		assert.Len(t, store, 0)
	})

	t.Run("should handle duplicate entries", func(t *testing.T) {
		store := make(StatusesStore)
		report := &Report{
			{
				UserId:    "michael.bolton@initech.com",
				ContentId: "bs_adg02_a23_enus",
				Status:    "Started",
			},
			{
				UserId:    "michael.bolton@initech.com",
				ContentId: "bs_adg02_a23_enus",
				Status:    "Completed",
			},
		}

		err := store.Load(ctx, report)
		require.NoError(t, err)

		course1Users := store.Get("bs_adg02_a23_enus")
		assert.Len(t, course1Users, 1)
		assert.Equal(t, "completed", course1Users["michael.bolton@initech.com"])
	})
}

func TestStatusesStoreGet(t *testing.T) {
	store := make(StatusesStore)

	store["bs_adg02_a23_enus"] = map[string]string{
		"michael.bolton@initech.com": "completed",
		"milton.waddams@initech.com": "in_progress",
		"peter.gibbons@initech.com":  "no_status_reported",
		"bill.lumbergh@initech.com":  "status_undefined",
	}

	t.Run("should return existing course mappings", func(t *testing.T) {
		users := store.Get("bs_adg02_a23_enus")
		assert.Len(t, users, 4)
		assert.Equal(t, "completed", users["michael.bolton@initech.com"])
		assert.Equal(t, "in_progress", users["milton.waddams@initech.com"])
		assert.Equal(t, "no_status_reported", users["peter.gibbons@initech.com"])
		assert.Equal(t, "status_undefined", users["bill.lumbergh@initech.com"])
	})

	t.Run("should return nil for non-existent course", func(t *testing.T) {
		users := store.Get("nonexistent_course")
		assert.Nil(t, users)
	})
}

func TestToStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected string
	}{
		{"completed status", "Completed", "completed"},
		{"achieved status", "Achieved", "completed"},
		{"listened status", "Listened", "completed"},
		{"read status", "Read", "completed"},
		{"watched status", "Watched", "completed"},
		{"started status", "Started", "in_progress"},
		{"active status", "Active", "in_progress"},
		{"empty status", "", "no_status_reported"},
		{"unknown status", "SomeRandomStatus", "status_undefined"},
		{"another unknown status", "InvalidStatus", "status_undefined"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toStatus(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStatusesStoreIntegration(t *testing.T) {
	ctx := context.Background()

	t.Run("should handle realistic report data", func(t *testing.T) {
		store := make(StatusesStore)
		report := &Report{
			{
				UserId:        "michael.bolton@initech.com",
				FirstName:     "Michael",
				LastName:      "Bolton",
				EmailAddress:  "michael.bolton@initech.com",
				ContentId:     "bs_adg02_a23_enus",
				ContentTitle:  "Case Studies: Successful Data Privacy Implementations",
				Status:        "Completed",
				CompletedDate: "2025-06-20T00:00:00.000Z",
			},
			{
				UserId:       "milton.waddams@initech.com",
				FirstName:    "Milton",
				LastName:     "Waddams",
				EmailAddress: "milton.waddams@initech.com",
				ContentId:    "bs_adg02_a23_enus",
				ContentTitle: "Case Studies: Successful Data Privacy Implementations",
				Status:       "Started",
				FirstAccess:  "2025-06-20T15:54:14.704Z",
			},
			{
				UserId:       "michael.bolton@initech.com",
				FirstName:    "Michael",
				LastName:     "Bolton",
				EmailAddress: "michael.bolton@initech.com",
				ContentId:    "another_course_id",
				ContentTitle: "Another Course",
				Status:       "Active",
				LastAccess:   "2025-06-20T16:00:43.775Z",
			},
			{
				UserId:       "peter.gibbons@initech.com",
				FirstName:    "Peter",
				LastName:     "Gibbons",
				EmailAddress: "peter.gibbons@initech.com",
				ContentId:    "bs_adg02_a23_enus",
				ContentTitle: "Case Studies: Successful Data Privacy Implementations",
				Status:       "",
			},
			{
				UserId:       "bill.lumbergh@initech.com",
				FirstName:    "Bill",
				LastName:     "Lumbergh",
				EmailAddress: "bill.lumbergh@initech.com",
				ContentId:    "bs_adg02_a23_enus",
				ContentTitle: "Case Studies: Successful Data Privacy Implementations",
				Status:       "SomeUnknownStatus",
			},
		}

		err := store.Load(ctx, report)
		require.NoError(t, err)

		assert.Len(t, store, 2)

		course1 := store.Get("bs_adg02_a23_enus")
		assert.Len(t, course1, 4)
		assert.Equal(t, "completed", course1["michael.bolton@initech.com"])
		assert.Equal(t, "in_progress", course1["milton.waddams@initech.com"])
		assert.Equal(t, "no_status_reported", course1["peter.gibbons@initech.com"])
		assert.Equal(t, "status_undefined", course1["bill.lumbergh@initech.com"])

		course2 := store.Get("another_course_id")
		assert.Len(t, course2, 1)
		assert.Equal(t, "in_progress", course2["michael.bolton@initech.com"])
	})
}
