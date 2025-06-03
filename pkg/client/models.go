package client

import "time"

type Course struct {
	Id   string `json:"id"`   // Will use contentId from report
	Name string `json:"name"` // Will use contentTitle from report
	// UUID string `json:"uuid"` // Will use contentUuid from report (commented for now)
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
	ContentUUID   string `json:"contentUiud"`
	ContentTitle  string `json:"contentTitle"`
	ContentType   string `json:"contentType"`
	Status        string `json:"status"`
	CompletedDate string `json:"completedDate"`
	FirstAccess   string `json:"firstAccess"`
	LastAccess    string `json:"lastAccess"`
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
	Id        string `json:"id"`        // Will use userId from report
	Email     string `json:"email"`     // Will use emailAddress from report
	FirstName string `json:"firstName"` // Will use firstName from report
	LastName  string `json:"lastName"`  // Will use lastName from report
}
