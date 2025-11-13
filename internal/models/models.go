package models

// Personnel represents an individual that can be assigned to a shift.
type Personnel struct {
	ID   int    `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

// PersonnelSchedule represents an invidiual and their assigned shift schedules.
type PersonnelSchedule struct {
	Personnel *Personnel       `json:"personnel" db:"personnel"`
	Schedules []*ShiftSchedule `json:"schedules" db:"schedules"`
}

// ScheduleType defines the template or category of a schedule.
type ScheduleType struct {
	ID          int     `json:"id" db:"id"`
	Name        string  `json:"name" db:"name"`
	Description *string `json:"description,omitempty" db:"description"`
}

// ScheduleTypePersonnel defines which personnel are eligible to participate in which schedule types.
type ScheduleTypePersonnel struct {
	ScheduleType ScheduleType `json:"schedule_type" db:"schedule_type"`
	Personnel    []*Personnel `json:"personnel" db:"personnel"`
}

// ShiftSchedule represents a concrete schedule instance (e.g., a specific week).
type ShiftSchedule struct {
	ID             int     `json:"id" db:"id"`
	Name           string  `json:"name" db:"name"`
	WeekNumber     int     `json:"weeknumber" db:"weeknumber"`
	Assignee       *int    `json:"assignee,omitempty" db:"assignee"`
	Substitute     *int    `json:"substitute,omitempty" db:"substitute"`
	Comment        *string `json:"comment,omitempty" db:"comment"`
	Accepted       bool    `json:"accepted" db:"accepted"`
	ScheduleTypeID int     `json:"schedule_type_id" db:"schedule_type_id"`
}
