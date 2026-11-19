package domain

import (
	"time"
)

const (
	// TODO: make these configurable, e.g. via a Calendar struct.
	handoverHour    = 8
	dayShiftEndHour = 15
	dayShiftEndMin  = 30
	daysInWeek      = 7
)

// Calendar knows when weeks start, which days are red days and when shifts end. Overrides maps a date (UTC midnight)
// to whether it is a red day, and takes precedence over weekends and the computed public holidays.
type Calendar struct {
	Loc       *time.Location
	Overrides map[time.Time]bool
}

// DateOf returns the calendar date of t in the calendar's location as UTC midnight.
func (c *Calendar) DateOf(t time.Time) time.Time {
	y, m, d := t.In(c.Loc).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// WeekMonday returns the Monday date of the week t is in. Weeks run from Monday 08:00 to the next Monday 08:00.
func (c *Calendar) WeekMonday(t time.Time) time.Time {
	local := t.In(c.Loc)
	date := c.DateOf(t)
	daysSinceMonday := (int(date.Weekday()) + daysInWeek - int(time.Monday)) % daysInWeek
	if daysSinceMonday == 0 && local.Hour() < handoverHour {
		daysSinceMonday = daysInWeek
	}
	return date.AddDate(0, 0, -daysSinceMonday)
}

// WeekBounds returns the instants at which the week starting on monday begins and ends.
func (c *Calendar) WeekBounds(monday time.Time) (time.Time, time.Time) {
	return c.at(monday, handoverHour, 0), c.at(monday.AddDate(0, 0, daysInWeek), handoverHour, 0)
}

// AddWeeks moves a Monday date n weeks.
func AddWeeks(monday time.Time, n int) time.Time {
	return monday.AddDate(0, 0, n*daysInWeek)
}

// IsRedDay reports whether the date is a weekend or a public holiday, subject to overrides.
func (c *Calendar) IsRedDay(date time.Time) bool {
	if red, ok := c.Overrides[date]; ok {
		return red
	}
	if wd := date.Weekday(); wd == time.Saturday || wd == time.Sunday {
		return true
	}
	_, ok := NorwegianHolidays(date.Year())[date]
	return ok
}

// RestOfShiftEnd returns when "the rest of the day" ends for a position of the given kind, seen from now:
// 15:30 for the day shift, the next 08:00 for on-call. The result is capped at weekEnd.
func (c *Calendar) RestOfShiftEnd(kind ShiftKind, now, weekEnd time.Time) (time.Time, error) {
	date := c.DateOf(now)
	var end time.Time

	switch kind {
	case KindDayShift:
		end = c.at(date, dayShiftEndHour, dayShiftEndMin)
		if c.IsRedDay(date) || !now.Before(end) {
			return time.Time{}, ErrNoShiftRemaining
		}
	case KindOnCall:
		end = c.at(date, handoverHour, 0)
		if !now.Before(end) {
			end = c.at(date.AddDate(0, 0, 1), handoverHour, 0)
		}
	default:
		return time.Time{}, ErrInvalid
	}

	if end.After(weekEnd) {
		end = weekEnd
	}
	return end, nil
}

func (c *Calendar) at(date time.Time, hour, minute int) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), hour, minute, 0, 0, c.Loc)
}

// NorwegianHolidays returns the public holidays of a year keyed by date (UTC midnight).
func NorwegianHolidays(year int) map[time.Time]string {
	date := func(m time.Month, d int) time.Time { return time.Date(year, m, d, 0, 0, 0, 0, time.UTC) }
	easter := easterSunday(year)
	offset := func(days int) time.Time { return easter.AddDate(0, 0, days) }

	return map[time.Time]string{
		date(time.January, 1):   "New Year's Day",
		offset(-3):              "Maundy Thursday",
		offset(-2):              "Good Friday",
		offset(0):               "Easter Sunday",
		offset(1):               "Easter Monday",
		date(time.May, 1):       "Labour Day",
		date(time.May, 17):      "Constitution Day",
		offset(39):              "Ascension Day",
		offset(49):              "Whit Sunday",
		offset(50):              "Whit Monday",
		date(time.December, 25): "Christmas Day",
		date(time.December, 26): "Boxing Day",
	}
}

// easterSunday uses the anonymous Gregorian algorithm.
func easterSunday(year int) time.Time {
	a := year % 19
	b := year / 100
	c := year % 100
	d := b / 4
	e := b % 4
	f := (b + 8) / 25
	g := (b - f + 1) / 3
	h := (19*a + b - d - g + 15) % 30
	i := c / 4
	k := c % 4
	l := (32 + 2*e + 2*i - h - k) % 7
	m := (a + 11*h + 22*l) / 451
	month := (h + l - 7*m + 114) / 31
	day := (h+l-7*m+114)%31 + 1
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}
