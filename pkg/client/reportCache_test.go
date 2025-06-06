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
				UserId:      "user1",
				ContentId:   "course1",
				Status:      "Completed",
				EmailAddress: "user1@example.com",
			},
			{
				UserId:      "user2",
				ContentId:   "course1",
				Status:      "Started",
				EmailAddress: "user2@example.com",
			},
			{
				UserId:      "user1",
				ContentId:   "course2",
				Status:      "Active",
				EmailAddress: "user1@example.com",
			},
		}

		err := store.Load(ctx, report)
		require.NoError(t, err)

		// Should have 2 courses
		assert.Len(t, store, 2)

		// Check course1 mappings
		course1Users := store.Get("course1")
		assert.Len(t, course1Users, 2)
		assert.Equal(t, "completed", course1Users["user1"])
		assert.Equal(t, "in_progress", course1Users["user2"])

		// Check course2 mappings
		course2Users := store.Get("course2")
		assert.Len(t, course2Users, 1)
		assert.Equal(t, "in_progress", course2Users["user1"])
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
				UserId:    "user1",
				ContentId: "course1",
				Status:    "Started",
			},
			{
				UserId:    "user1",
				ContentId: "course1",
				Status:    "Completed", // Should overwrite the first entry
			},
		}

		err := store.Load(ctx, report)
		require.NoError(t, err)

		course1Users := store.Get("course1")
		assert.Len(t, course1Users, 1)
		assert.Equal(t, "completed", course1Users["user1"]) // Should be the latest status
	})
}

func TestStatusesStoreGet(t *testing.T) {
	store := make(StatusesStore)
	
	// Manually populate store
	store["course1"] = map[string]string{
		"user1": "completed",
		"user2": "in_progress",
	}

	t.Run("should return existing course mappings", func(t *testing.T) {
		users := store.Get("course1")
		assert.Len(t, users, 2)
		assert.Equal(t, "completed", users["user1"])
		assert.Equal(t, "in_progress", users["user2"])
	})

	t.Run("should return nil for non-existent course", func(t *testing.T) {
		users := store.Get("non-existent")
		assert.Nil(t, users)
	})
}

func TestToStatus(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		{"Completed", "completed", "completed status"},
		{"Achieved", "completed", "achieved status"},
		{"Listened", "completed", "listened status"},
		{"Read", "completed", "read status"},
		{"Watched", "completed", "watched status"},
		{"Started", "in_progress", "started status"},
		{"Active", "in_progress", "active status"},
		{"Unknown", "in_progress", "unknown status defaults to in_progress"},
		{"", "in_progress", "empty status defaults to in_progress"},
		{"SomeRandomStatus", "in_progress", "random status defaults to in_progress"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := toStatus(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestStatusesStoreIntegration(t *testing.T) {
	ctx := context.Background()
	
	t.Run("should handle realistic report data", func(t *testing.T) {
		store := make(StatusesStore)
		report := &Report{
			{
				UserId:        "user1",
				FirstName:     "John",
				LastName:      "Doe",
				EmailAddress:  "john@example.com",
				ContentId:     "course1",
				ContentTitle:  "Introduction to Go",
				Status:        "Completed",
				CompletedDate: "2023-01-15",
			},
			{
				UserId:        "user2",
				FirstName:     "Jane",
				LastName:      "Smith",
				EmailAddress:  "jane@example.com",
				ContentId:     "course1",
				ContentTitle:  "Introduction to Go",
				Status:        "Started",
				FirstAccess:   "2023-01-10",
			},
			{
				UserId:        "user1",
				FirstName:     "John",
				LastName:      "Doe",
				EmailAddress:  "john@example.com",
				ContentId:     "course2",
				ContentTitle:  "Advanced Go Patterns",
				Status:        "Active",
				LastAccess:    "2023-01-20",
			},
		}

		err := store.Load(ctx, report)
		require.NoError(t, err)

		// Verify the loaded data
		assert.Len(t, store, 2) // 2 courses

		// Course 1 should have 2 users
		course1 := store.Get("course1")
		assert.Len(t, course1, 2)
		assert.Equal(t, "completed", course1["user1"])
		assert.Equal(t, "in_progress", course1["user2"])

		// Course 2 should have 1 user
		course2 := store.Get("course2")
		assert.Len(t, course2, 1)
		assert.Equal(t, "in_progress", course2["user1"])
	})
}