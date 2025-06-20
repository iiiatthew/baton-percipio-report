package client

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUserModel(t *testing.T) {
	user := User{
		Id:        "a77840ca-ea10-4da8-b64f-bddf714c47a0",
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
	}

	assert.Equal(t, "a77840ca-ea10-4da8-b64f-bddf714c47a0", user.Id)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "John", user.FirstName)
	assert.Equal(t, "Doe", user.LastName)
}

func TestCourseModel(t *testing.T) {
	course := Course{
		Id:          "1a3a3f54-b601-4d45-a234-038c980ee20f",
		CourseTitle: "Introduction to Go",
	}

	assert.Equal(t, "1a3a3f54-b601-4d45-a234-038c980ee20f", course.Id)
	assert.Equal(t, "Introduction to Go", course.CourseTitle)
}

func TestReportEntry(t *testing.T) {
	entry := ReportEntry{
		UserUUID:      "a77840ca-ea10-4da8-b64f-bddf714c47a0",
		FirstName:     "John",
		LastName:      "Doe",
		EmailAddress:  "john@example.com",
		ContentUUID:   "1a3a3f54-b601-4d45-a234-038c980ee20f",
		ContentTitle:  "Test Course",
		ContentType:   "Course",
		Status:        "Completed",
		CompletedDate: "2023-01-15",
		FirstAccess:   "2023-01-10",
		LastAccess:    "2023-01-14",
	}

	assert.Equal(t, "a77840ca-ea10-4da8-b64f-bddf714c47a0", entry.UserUUID)
	assert.Equal(t, "John", entry.FirstName)
	assert.Equal(t, "Doe", entry.LastName)
	assert.Equal(t, "john@example.com", entry.EmailAddress)
	assert.Equal(t, "Introduction to Go", entry.ContentTitle)
	assert.Equal(t, "1a3a3f54-b601-4d45-a234-038c980ee20f", entry.ContentUUID)
	assert.Equal(t, "Course", entry.ContentType)
	assert.Equal(t, "Completed", entry.Status)
	assert.Equal(t, "2023-01-15", entry.CompletedDate)
	assert.Equal(t, "2023-01-10", entry.FirstAccess)
	assert.Equal(t, "2023-01-14", entry.LastAccess)
}

func TestReportStatus(t *testing.T) {
	status := ReportStatus{
		Id:     "bed2ffcb-9822-4dd1-bc8b-3eda6e5eff41",
		Status: "COMPLETED",
	}

	assert.Equal(t, "bed2ffcb-9822-4dd1-bc8b-3eda6e5eff41", status.Id)
	assert.Equal(t, "COMPLETED", status.Status)
}

func TestReportConfigurations(t *testing.T) {
	config := ReportConfigurations{
		ContentType: "Course,Assessment",
		End:         time.Now(),
		Start:       time.Now().Add(-time.Hour * 24 * 30),
		Sort:        &ReportSort{Field: "completedDate", Order: "desc"},
	}

	assert.Equal(t, "Course,Assessment", config.ContentType)
	assert.Equal(t, time.Now(), config.End)
	assert.Equal(t, time.Now().Add(-time.Hour*24*30), config.Start)
	assert.Equal(t, "completedDate", config.Sort.Field)
	assert.Equal(t, "desc", config.Sort.Order)
}

func TestReportType(t *testing.T) {
	// Test that Report is a slice of ReportEntry
	report := Report{
		{
			UserUUID:    "a77840ca-ea10-4da8-b64f-bddf714c47a0",
			ContentUUID: "1a3a3f54-b601-4d45-a234-038c980ee20f",
			Status:      "Completed",
		},
		{
			UserUUID:    "a77840ca-ea10-4da8-b64f-bddf714c47a0",
			ContentUUID: "1a3a3f54-b601-4d45-a234-038c980ee20f",
			Status:      "Started",
		},
	}

	assert.Len(t, report, 2)
	assert.Equal(t, "a77840ca-ea10-4da8-b64f-bddf714c47a0", report[0].UserUUID)
	assert.Equal(t, "1a3a3f54-b601-4d45-a234-038c980ee20f", report[0].ContentUUID)
	assert.Equal(t, "Completed", report[0].Status)
	assert.Equal(t, "a77840ca-ea10-4da8-b64f-bddf714c47a0", report[1].UserUUID)
	assert.Equal(t, "1a3a3f54-b601-4d45-a234-038c980ee20f", report[1].ContentUUID)
	assert.Equal(t, "Started", report[1].Status)
}
