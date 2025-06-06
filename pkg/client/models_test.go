package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserModel(t *testing.T) {
	user := User{
		Id:        "user123",
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
	}

	assert.Equal(t, "user123", user.Id)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "John", user.FirstName)
	assert.Equal(t, "Doe", user.LastName)
}

func TestCourseModel(t *testing.T) {
	course := Course{
		Id:   "course123",
		Name: "Introduction to Go",
	}

	assert.Equal(t, "course123", course.Id)
	assert.Equal(t, "Introduction to Go", course.Name)
}

func TestReportEntry(t *testing.T) {
	entry := ReportEntry{
		UserId:        "user123",
		FirstName:     "John",
		LastName:      "Doe",
		EmailAddress:  "john@example.com",
		ContentId:     "course123",
		ContentTitle:  "Test Course",
		ContentUUID:   "uuid-123",
		ContentType:   "Course",
		Status:        "Completed",
		CompletedDate: "2023-01-15",
		FirstAccess:   "2023-01-10",
		LastAccess:    "2023-01-14",
	}

	assert.Equal(t, "user123", entry.UserId)
	assert.Equal(t, "John", entry.FirstName)
	assert.Equal(t, "Doe", entry.LastName)
	assert.Equal(t, "john@example.com", entry.EmailAddress)
	assert.Equal(t, "course123", entry.ContentId)
	assert.Equal(t, "Test Course", entry.ContentTitle)
	assert.Equal(t, "uuid-123", entry.ContentUUID)
	assert.Equal(t, "Course", entry.ContentType)
	assert.Equal(t, "Completed", entry.Status)
	assert.Equal(t, "2023-01-15", entry.CompletedDate)
	assert.Equal(t, "2023-01-10", entry.FirstAccess)
	assert.Equal(t, "2023-01-14", entry.LastAccess)
}

func TestReportStatus(t *testing.T) {
	status := ReportStatus{
		Id:     "report123",
		Status: "COMPLETED",
	}

	assert.Equal(t, "report123", status.Id)
	assert.Equal(t, "COMPLETED", status.Status)
}

func TestReportConfigurations(t *testing.T) {
	config := ReportConfigurations{
		ContentType: "Course,Assessment",
	}

	assert.Equal(t, "Course,Assessment", config.ContentType)
}

func TestReportType(t *testing.T) {
	// Test that Report is a slice of ReportEntry
	report := Report{
		{
			UserId:    "user1",
			ContentId: "course1",
			Status:    "Completed",
		},
		{
			UserId:    "user2",
			ContentId: "course2",
			Status:    "Started",
		},
	}

	assert.Len(t, report, 2)
	assert.Equal(t, "user1", report[0].UserId)
	assert.Equal(t, "course1", report[0].ContentId)
	assert.Equal(t, "Completed", report[0].Status)
	assert.Equal(t, "user2", report[1].UserId)
	assert.Equal(t, "course2", report[1].ContentId)
	assert.Equal(t, "Started", report[1].Status)
}