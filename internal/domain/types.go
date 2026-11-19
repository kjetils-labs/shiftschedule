// Package domain holds the storage-independent business rules: shift types, week handling, holidays and rotation generation.
package domain

import (
	"errors"
	"time"
)

var (
	ErrNotFound          = errors.New("not found")
	ErrForbidden         = errors.New("forbidden")
	ErrInvalid           = errors.New("invalid input")
	ErrConflict          = errors.New("conflict")
	ErrNoShiftRemaining  = errors.New("no shift remaining")
	ErrNotTeamMember     = errors.New("person is not a member of the position's team")
	ErrWeekAlreadyPassed = errors.New("week has already passed")
)

// ShiftKind is the type of a position.
type ShiftKind string

const (
	// KindDayShift works 08:00-15:30 on working weekdays.
	KindDayShift ShiftKind = "day_shift"
	// KindOnCall covers everything the day shift does not: 15:30-08:00 on working days and around the clock on weekends and red days.
	KindOnCall ShiftKind = "on_call"
)

// Valid reports whether k is a known kind.
func (k ShiftKind) Valid() bool {
	return k == KindDayShift || k == KindOnCall
}

type Personnel struct {
	Name    string   `json:"name"`
	Email   string   `json:"email"`
	Groups  []string `json:"groups"`
	ID      int64    `json:"id"`
	IsAdmin bool     `json:"is_admin"`
}

// PersonRef is the public, minimal representation of a person.
type PersonRef struct {
	Name string `json:"name"`
	ID   int64  `json:"id"`
}

// Team is a set of people and the positions they man. Members come from the OIDC Group, and the Manager is in charge of the team's shifts.
type Team struct {
	ManagerID *int64 `json:"manager_id"`
	Name      string `json:"name"`
	Group     string `json:"group"`
	ID        int64  `json:"id"`
}

// Position is a named slot in a team. Everyone in the team may man or cover it.
type Position struct {
	ManagerID *int64    `json:"manager_id"`
	TeamName  string    `json:"team_name"`
	Name      string    `json:"name"`
	Kind      ShiftKind `json:"kind"`
	ID        int64     `json:"id"`
	TeamID    int64     `json:"team_id"`
}

// Assignment is who holds a position between StartsAt and EndsAt. A nil Assignee means the period is not filled.
type Assignment struct {
	StartsAt   time.Time
	EndsAt     time.Time
	Assignee   *PersonRef
	Covers     []Cover
	ID         int64
	PositionID int64
}

// Cover is someone temporarily taking over an assignment between StartsAt and EndsAt.
type Cover struct {
	StartsAt     time.Time
	EndsAt       time.Time
	Reason       string
	CoveredBy    PersonRef
	ID           int64
	AssignmentID int64
}
