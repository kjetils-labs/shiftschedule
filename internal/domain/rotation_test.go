package domain_test

import (
	"testing"

	"github.com/shiftschedule/internal/domain"
	"github.com/stretchr/testify/assert"
)

func assigneesByPosition(planned []domain.PlannedAssignment, positionID int64) []int64 {
	var out []int64
	for _, p := range planned {
		if p.PositionID == positionID {
			out = append(out, p.AssigneeID)
		}
	}
	return out
}

func TestMondaysOfYear(t *testing.T) {
	t.Parallel()
	mondays := domain.MondaysOfYear(2026)

	assert.Equal(t, utcDate(2026, 1, 5), mondays[0])
	assert.Equal(t, utcDate(2026, 12, 28), mondays[len(mondays)-1])
	assert.Len(t, mondays, 52)
}

func TestGenerateRotationRoundRobinContinuesAfterLast(t *testing.T) {
	t.Parallel()
	positions := []domain.RotationPosition{
		{PositionID: 1, Kind: domain.KindOnCall, Pool: []int64{10, 11, 12}, Last: 11},
	}

	planned := domain.GenerateRotation(positions, domain.MondaysOfYear(2026)[:4])

	assert.Equal(t, []int64{12, 10, 11, 12}, assigneesByPosition(planned, 1))
}

func TestGenerateRotationDayShiftAvoidsOnCallAndSkippedStaysInLine(t *testing.T) {
	t.Parallel()
	positions := []domain.RotationPosition{
		{PositionID: 1, Kind: domain.KindDayShift, Pool: []int64{10, 11, 12}},
		{PositionID: 2, Kind: domain.KindOnCall, Pool: []int64{10, 11, 12}},
	}

	planned := domain.GenerateRotation(positions, domain.MondaysOfYear(2026)[:3])
	onCall := assigneesByPosition(planned, 2)
	day := assigneesByPosition(planned, 1)

	assert.Equal(t, []int64{10, 11, 12}, onCall)
	assert.Equal(t, []int64{11, 10, 11}, day, "10 is skipped in week 1 because they are on-call, and gets week 2 instead")
	for i := range day {
		assert.NotEqual(t, onCall[i], day[i])
	}
}

func TestGenerateRotationKeepsOnePersonOnAllOnCallPositions(t *testing.T) {
	t.Parallel()
	positions := []domain.RotationPosition{
		{PositionID: 1, Kind: domain.KindOnCall, Pool: []int64{10, 11, 12}},
		{PositionID: 2, Kind: domain.KindOnCall, Pool: []int64{12, 10, 11}},
		{PositionID: 3, Kind: domain.KindOnCall, Pool: []int64{20, 21}},
	}

	planned := domain.GenerateRotation(positions, domain.MondaysOfYear(2026)[:6])

	assert.Equal(t, assigneesByPosition(planned, 1), assigneesByPosition(planned, 2))
	assert.Equal(t, []int64{20, 21, 20, 21, 20, 21}, assigneesByPosition(planned, 3), "not in the pool of the others, plain round robin")
}

func TestGenerateRotationEmptyPool(t *testing.T) {
	t.Parallel()
	planned := domain.GenerateRotation([]domain.RotationPosition{{PositionID: 1, Kind: domain.KindDayShift}}, domain.MondaysOfYear(2026))
	assert.Empty(t, planned)
}
