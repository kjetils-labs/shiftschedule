package models

type Personnel struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type SchedulePersonnel struct {
	ScheduleName string       `json:"schedule_name"`
	Personnel    []*Personnel `json:"personnel"`
}

type ShiftSchedule struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	WeekNumber int     `json:"week_number"`
	Assignee   int     `json:"assignee"`
	Substitute *int    `json:"substitute"`
	Comment    *string `json:"comment"`
	Accepted   bool    `json:"accepted"`
}
