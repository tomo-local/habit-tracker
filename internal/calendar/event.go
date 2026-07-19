package calendar

import "time"

type Event struct {
	Title string
	Start time.Time
	End   *time.Time
}

type CalendarEntry struct {
	ID   string
	Name string
}
