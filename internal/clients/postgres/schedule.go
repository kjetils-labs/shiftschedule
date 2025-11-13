package postgres

import (
	"github.com/jackc/pgx/v5"
	"github.com/shiftschedule/internal/models"
)

func mapRowToShiftSchedule(rows pgx.Rows) (*models.ShiftSchedule, error) {
	var s models.ShiftSchedule
	if err := rows.Scan(
		&s.ID,
		&s.Name,
		&s.WeekNumber,
		&s.Assignee,
		&s.Substitute,
		&s.Comment,
		&s.Accepted,
	); err != nil {
		return &models.ShiftSchedule{}, err
	}
	return &s, nil
}
