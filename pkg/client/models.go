package client

import "time"

type Course struct {
	Id          string `json:"contentId"`
	CourseTitle string `json:"contentTitle"`
	ContentType string `json:"contentType"`
}

type Report []ReportEntry

type ReportConfigurations struct {
	ContentType string      `json:"contentType,omitempty"`
	End         time.Time   `json:"end,omitempty"`
	Start       time.Time   `json:"start,omitempty"`
	Sort        *ReportSort `json:"sort,omitempty"`
}

type ReportEntry struct {
	UserId        string `json:"userId"`
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	EmailAddress  string `json:"emailAddress"`
	ContentId     string `json:"contentId"`
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
	Id        string `json:"userId"`
	Email     string `json:"emailAddress"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}
