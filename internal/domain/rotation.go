package domain

import (
	"slices"
	"time"
)

// RotationPosition is the input for generating a rotation: the ordered team members and who held the position last.
type RotationPosition struct {
	Kind ShiftKind
	Pool []int64
	// Last is the most recent assignee before the first generated week, or 0 when there is none.
	Last       int64
	PositionID int64
}

// PlannedAssignment is a generated assignment.
type PlannedAssignment struct {
	WeekStart  time.Time
	PositionID int64
	AssigneeID int64
}

// PlannedSlot is a PlannedAssignment placed in time.
type PlannedSlot struct {
	StartsAt   time.Time
	EndsAt     time.Time
	PositionID int64
	AssigneeID int64
}

// MondaysOfYear returns every Monday that falls in the given year. A week belongs to the year its Monday is in.
func MondaysOfYear(year int) []time.Time {
	day := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	for day.Weekday() != time.Monday {
		day = day.AddDate(0, 0, 1)
	}

	var mondays []time.Time
	for ; day.Year() == year; day = AddWeeks(day, 1) {
		mondays = append(mondays, day)
	}
	return mondays
}

// GenerateRotation plans the weeks in mondays for the positions of one team using round robin through each pool, continuing after Last.
// Preferences, in order of priority:
//   - the day shift is not held by someone who holds an on-call position the same week,
//   - one person holds all on-call positions of the team the same week, if they are in all of those pools.
//
// A person skipped because of a preference stays first in line for the next week.
func GenerateRotation(positions []RotationPosition, mondays []time.Time) []PlannedAssignment {
	ordered := slices.Clone(positions)
	slices.SortStableFunc(ordered, func(a, b RotationPosition) int {
		return kindRank(a.Kind) - kindRank(b.Kind)
	})

	queues := make([][]int64, len(ordered))
	for i, p := range ordered {
		queues[i] = startQueue(p.Pool, p.Last)
	}

	var planned []PlannedAssignment
	for _, monday := range mondays {
		onCall := map[int64]bool{}
		var primaryOnCall int64

		for i, p := range ordered {
			if len(queues[i]) == 0 {
				continue
			}

			var picked int64
			switch {
			case p.Kind == KindOnCall && primaryOnCall != 0 && slices.Contains(queues[i], primaryOnCall):
				picked = primaryOnCall
			case p.Kind == KindDayShift:
				picked = firstWhere(queues[i], func(id int64) bool { return !onCall[id] })
			default:
				picked = queues[i][0]
			}

			queues[i] = rotateToBack(queues[i], picked)
			if p.Kind == KindOnCall {
				onCall[picked] = true
				if primaryOnCall == 0 {
					primaryOnCall = picked
				}
			}
			planned = append(planned, PlannedAssignment{WeekStart: monday, PositionID: p.PositionID, AssigneeID: picked})
		}
	}

	return planned
}

func kindRank(k ShiftKind) int {
	if k == KindOnCall {
		return 0
	}
	return 1
}

// startQueue rotates the pool so the person after last is first in line.
func startQueue(pool []int64, last int64) []int64 {
	i := slices.Index(pool, last)
	if last == 0 || i < 0 {
		return slices.Clone(pool)
	}
	return append(slices.Clone(pool[i+1:]), pool[:i+1]...)
}

// firstWhere returns the first id matching ok, or the first id when none match.
func firstWhere(queue []int64, ok func(int64) bool) int64 {
	if i := slices.IndexFunc(queue, ok); i >= 0 {
		return queue[i]
	}
	return queue[0]
}

func rotateToBack(queue []int64, id int64) []int64 {
	i := slices.Index(queue, id)
	out := make([]int64, 0, len(queue))
	out = append(out, queue[:i]...)
	out = append(out, queue[i+1:]...)
	return append(out, id)
}
