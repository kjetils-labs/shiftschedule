package domain_test

import (
	"testing"
	"time"

	"github.com/shiftschedule/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCalendar(t *testing.T) *domain.Calendar {
	t.Helper()
	loc, err := time.LoadLocation("Europe/Oslo")
	require.NoError(t, err)
	return &domain.Calendar{Loc: loc, Overrides: map[time.Time]bool{}}
}

func utcDate(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func TestWeekMondayHandoverIsMonday0800(t *testing.T) {
	t.Parallel()
	cal := newCalendar(t)

	tests := map[string]struct {
		now  time.Time
		want time.Time
	}{
		"monday before handover is the previous week": {time.Date(2026, 11, 16, 7, 59, 0, 0, cal.Loc), utcDate(2026, 11, 9)},
		"monday at handover is the new week":          {time.Date(2026, 11, 16, 8, 0, 0, 0, cal.Loc), utcDate(2026, 11, 16)},
		"sunday night":                                {time.Date(2026, 11, 15, 23, 0, 0, 0, cal.Loc), utcDate(2026, 11, 9)},
		"wednesday":                                   {time.Date(2026, 11, 18, 8, 34, 0, 0, cal.Loc), utcDate(2026, 11, 16)},
		"given in utc but monday 00:30 in oslo":       {time.Date(2026, 11, 15, 23, 30, 0, 0, time.UTC), utcDate(2026, 11, 9)},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, cal.WeekMonday(tt.now))
		})
	}
}

func TestWeekBoundsAcrossDST(t *testing.T) {
	t.Parallel()
	cal := newCalendar(t)

	// Clocks go back on 2026-10-25, so the week is one hour longer than 168h.
	start, end := cal.WeekBounds(utcDate(2026, 10, 19))
	assert.Equal(t, 8, start.In(cal.Loc).Hour())
	assert.Equal(t, 8, end.In(cal.Loc).Hour())
	assert.Equal(t, 169*time.Hour, end.Sub(start))
}

func TestNorwegianHolidays(t *testing.T) {
	t.Parallel()
	holidays := domain.NorwegianHolidays(2026)

	assert.Contains(t, holidays, utcDate(2026, 4, 3), "good friday")
	assert.Contains(t, holidays, utcDate(2026, 4, 6), "easter monday")
	assert.Contains(t, holidays, utcDate(2026, 5, 14), "ascension")
	assert.Contains(t, holidays, utcDate(2026, 5, 17))
	assert.NotContains(t, holidays, utcDate(2026, 4, 7))
}

func TestIsRedDayOverrides(t *testing.T) {
	t.Parallel()
	cal := newCalendar(t)

	assert.True(t, cal.IsRedDay(utcDate(2026, 11, 14)), "saturday")
	assert.False(t, cal.IsRedDay(utcDate(2026, 11, 18)), "wednesday")
	assert.True(t, cal.IsRedDay(utcDate(2026, 12, 25)), "christmas")

	cal.Overrides[utcDate(2026, 11, 18)] = true
	cal.Overrides[utcDate(2026, 12, 25)] = false
	assert.True(t, cal.IsRedDay(utcDate(2026, 11, 18)))
	assert.False(t, cal.IsRedDay(utcDate(2026, 12, 25)))
}

func TestRestOfShiftEnd(t *testing.T) {
	t.Parallel()
	cal := newCalendar(t)
	weekEnd := time.Date(2026, 11, 23, 8, 0, 0, 0, cal.Loc)
	at := func(d, h, m int) time.Time { return time.Date(2026, 11, d, h, m, 0, 0, cal.Loc) }

	end, err := cal.RestOfShiftEnd(domain.KindDayShift, at(18, 9, 0), weekEnd)
	require.NoError(t, err)
	assert.Equal(t, at(18, 15, 30), end)

	_, err = cal.RestOfShiftEnd(domain.KindDayShift, at(18, 16, 0), weekEnd)
	require.ErrorIs(t, err, domain.ErrNoShiftRemaining)

	_, err = cal.RestOfShiftEnd(domain.KindDayShift, at(21, 9, 0), weekEnd)
	require.ErrorIs(t, err, domain.ErrNoShiftRemaining, "saturday has no day shift")

	end, err = cal.RestOfShiftEnd(domain.KindOnCall, at(18, 17, 0), weekEnd)
	require.NoError(t, err)
	assert.Equal(t, at(19, 8, 0), end)

	end, err = cal.RestOfShiftEnd(domain.KindOnCall, at(19, 6, 0), weekEnd)
	require.NoError(t, err)
	assert.Equal(t, at(19, 8, 0), end)

	end, err = cal.RestOfShiftEnd(domain.KindOnCall, at(22, 12, 0), weekEnd)
	require.NoError(t, err)
	assert.Equal(t, weekEnd, end, "capped at the handover")
}
