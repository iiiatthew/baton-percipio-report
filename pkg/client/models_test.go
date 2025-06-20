package client

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUserModel(t *testing.T) {
	user := User{
		Id:        "michael.bolton@initech.com",
		Email:     "michael.bolton@initech.com",
		FirstName: "Michael",
		LastName:  "Bolton",
	}

	assert.Equal(t, "michael.bolton@initech.com", user.Id)
	assert.Equal(t, "michael.bolton@initech.com", user.Email)
	assert.Equal(t, "Michael", user.FirstName)
	assert.Equal(t, "Bolton", user.LastName)
}

func TestCourseModel(t *testing.T) {
	course := Course{
		Id:          "bs_adg02_a23_enus",
		CourseTitle: "Case Studies: Successful Data Privacy Implementations",
	}

	assert.Equal(t, "bs_adg02_a23_enus", course.Id)
	assert.Equal(t, "Case Studies: Successful Data Privacy Implementations", course.CourseTitle)
}

func TestReportEntry(t *testing.T) {
	entry := ReportEntry{
		UserId:        "michael.bolton@initech.com",
		FirstName:     "Michael",
		LastName:      "Bolton",
		EmailAddress:  "michael.bolton@initech.com",
		ContentId:     "bs_adg02_a23_enus",
		ContentTitle:  "Case Studies: Successful Data Privacy Implementations",
		ContentType:   "Course",
		Status:        "Completed",
		CompletedDate: "2025-06-20T00:00:00.000Z",
		FirstAccess:   "2025-06-20T16:00:39.770Z",
		LastAccess:    "2025-06-20T16:00:43.775Z",
	}

	assert.Equal(t, "michael.bolton@initech.com", entry.UserId)
	assert.Equal(t, "Michael", entry.FirstName)
	assert.Equal(t, "Bolton", entry.LastName)
	assert.Equal(t, "michael.bolton@initech.com", entry.EmailAddress)
	assert.Equal(t, "Case Studies: Successful Data Privacy Implementations", entry.ContentTitle)
	assert.Equal(t, "bs_adg02_a23_enus", entry.ContentId)
	assert.Equal(t, "Course", entry.ContentType)
	assert.Equal(t, "Completed", entry.Status)
	assert.Equal(t, "2025-06-20T00:00:00.000Z", entry.CompletedDate)
	assert.Equal(t, "2025-06-20T16:00:39.770Z", entry.FirstAccess)
	assert.Equal(t, "2025-06-20T16:00:43.775Z", entry.LastAccess)
}

func TestReportEntryEmptyStatus(t *testing.T) {
	entry := ReportEntry{
		UserId:       "peter.gibbons@initech.com",
		FirstName:    "Peter",
		LastName:     "Gibbons",
		EmailAddress: "peter.gibbons@initech.com",
		ContentId:    "bs_adg02_a23_enus",
		ContentTitle: "Case Studies: Successful Data Privacy Implementations",
		ContentType:  "Course",
		Status:       "",
	}

	assert.Equal(t, "peter.gibbons@initech.com", entry.UserId)
	assert.Equal(t, "Peter", entry.FirstName)
	assert.Equal(t, "Gibbons", entry.LastName)
	assert.Equal(t, "", entry.Status)
}

func TestReportEntryUndefinedStatus(t *testing.T) {
	entry := ReportEntry{
		UserId:       "bill.lumbergh@initech.com",
		FirstName:    "Bill",
		LastName:     "Lumbergh",
		EmailAddress: "bill.lumbergh@initech.com",
		ContentId:    "bs_adg02_a23_enus",
		ContentTitle: "Case Studies: Successful Data Privacy Implementations",
		ContentType:  "Course",
		Status:       "SomeWeirdStatus",
	}

	assert.Equal(t, "bill.lumbergh@initech.com", entry.UserId)
	assert.Equal(t, "Bill", entry.FirstName)
	assert.Equal(t, "Lumbergh", entry.LastName)
	assert.Equal(t, "SomeWeirdStatus", entry.Status)
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
	// Use fixed times to avoid timing issues in tests
	endTime := time.Date(2025, 6, 20, 16, 7, 22, 963536000, time.Local)
	startTime := endTime.Add(-time.Hour * 24 * 30)

	config := ReportConfigurations{
		ContentType: "Course,Assessment",
		End:         endTime,
		Start:       startTime,
		Sort:        &ReportSort{Field: "completedDate", Order: "desc"},
	}

	assert.Equal(t, "Course,Assessment", config.ContentType)
	assert.Equal(t, endTime, config.End)
	assert.Equal(t, startTime, config.Start)
	assert.Equal(t, "completedDate", config.Sort.Field)
	assert.Equal(t, "desc", config.Sort.Order)
}

func TestReportType(t *testing.T) {
	// Test that Report is a slice of ReportEntry
	report := Report{
		{
			UserId:    "michael.bolton@initech.com",
			ContentId: "bs_adg02_a23_enus",
			Status:    "Completed",
		},
		{
			UserId:    "milton.waddams@initech.com",
			ContentId: "bs_adg02_a23_enus",
			Status:    "Started",
		},
		{
			UserId:    "peter.gibbons@initech.com",
			ContentId: "bs_adg02_a23_enus",
			Status:    "",
		},
		{
			UserId:    "bill.lumbergh@initech.com",
			ContentId: "bs_adg02_a23_enus",
			Status:    "UnknownStatus",
		},
	}

	assert.Len(t, report, 4)
	assert.Equal(t, "michael.bolton@initech.com", report[0].UserId)
	assert.Equal(t, "bs_adg02_a23_enus", report[0].ContentId)
	assert.Equal(t, "Completed", report[0].Status)
	assert.Equal(t, "milton.waddams@initech.com", report[1].UserId)
	assert.Equal(t, "bs_adg02_a23_enus", report[1].ContentId)
	assert.Equal(t, "Started", report[1].Status)
	assert.Equal(t, "peter.gibbons@initech.com", report[2].UserId)
	assert.Equal(t, "", report[2].Status)
	assert.Equal(t, "bill.lumbergh@initech.com", report[3].UserId)
	assert.Equal(t, "UnknownStatus", report[3].Status)
}
