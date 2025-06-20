package client

import "time"

type Course struct {
	Id          string `json:"contentUuid"`  // Will use contentId from report
	CourseTitle string `json:"contentTitle"` // Will use contentTitle from report
}

type Report []ReportEntry

type ReportConfigurations struct {
	ContentType string      `json:"contentType,omitempty"`
	End         time.Time   `json:"end,omitempty"`
	Start       time.Time   `json:"start,omitempty"`
	Sort        *ReportSort `json:"sort,omitempty"`
}

type ReportEntry struct {
	UserUUID      string `json:"userUuid"`
	UserId        string `json:"userId"`
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	EmailAddress  string `json:"emailAddress"`
	ContentUUID   string `json:"contentUuid"`
	ContentTitle  string `json:"contentTitle"`
	ContentType   string `json:"contentType"`
	Status        string `json:"status"`
	CompletedDate string `json:"completedDate,omitempty"`
	FirstAccess   string `json:"firstAccess,omitempty"`
	LastAccess    string `json:"lastAccess,omitempty"`
}

type ReportSort struct {
	Field string `json:"field,omitempty"`
	Order string `json:"order,omitempty"`
}

type ReportStatus struct {
	Id     string `json:"id"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type User struct {
	Id        string `json:"userUuid"`     // Will use userId from report
	LoginID   string `json:"userId"`       // Will use userId from report
	Email     string `json:"emailAddress"` // Will use emailAddress from report
	FirstName string `json:"firstName"`    // Will use firstName from report
	LastName  string `json:"lastName"`     // Will use lastName from report
}
