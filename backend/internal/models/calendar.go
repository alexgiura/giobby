package models

// Calendar is a user-visible calendar.
type Calendar struct {
	ID    int32  `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Color string `json:"color,omitempty"`
}

// CalendarEvent matches swagger CalendarEventApi (+ id).
type CalendarEvent struct {
	ID         int32  `json:"id,omitempty"`
	Event      string `json:"event,omitempty"`
	Place      string `json:"place,omitempty"`
	StartDate  *int64 `json:"startDate,omitempty"`
	EndDate    *int64 `json:"endDate,omitempty"`
	Color      string `json:"color,omitempty"`
	IDCalendar int32  `json:"idCalendar,omitempty"`
	Privacy    string `json:"privacy,omitempty"`
}

type CalendarListQuery struct {
	ListQuery
	Name string
}

type CalendarEventListQuery struct {
	ListQuery
	StartDateMillis string
	EndDateMillis   string
	FreeText        string
}
