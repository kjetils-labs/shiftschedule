package models

// Personnel represents an individual that can be assigned to a shift.

// CREATE TABLE personnel (
//
//	id SERIAL PRIMARY KEY,
//	name TEXT NOT NULL
//
// ):
// .
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

// ScheduleRelation defines which personnel are eligible to participate in which schedule types.
type ScheduleRelation struct {
	ScheduleType ScheduleType `json:"schedule_type" db:"schedule_type"`
	Personnel    []*Personnel `json:"personnel" db:"personnel"`
}

// ScheduleRelationID defines which personnel are eligible to participate in which schedule types.
type ScheduleRelationID struct {
	ScheduleType int `json:"schedule_type" db:"schedule_type"`
	Personnel    int `json:"personnel" db:"personnel"`
}

// ShiftSchedule represents a concrete schedule instance (e.g., a specific week).
// CREATE TABLE shiftschedule (
//
//	id SERIAL PRIMARY KEY,
//	schedule_type_id INT REFERENCES schedule_type(id)
//	name TEXT NOT NULL,
//	weeknumber INTEGER NOT NULL,
//	assignee INTEGER REFERENCES personnel(id),
//	substitute INTEGER REFERENCES personnel(id),
//	comment TEXT,
//	accepted BOOLEAN DEFAULT FALSE
//
// );
// .
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
