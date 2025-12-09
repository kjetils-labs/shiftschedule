package postgres

import (
	"github.com/jackc/pgx/v5"
	"github.com/shiftschedule/internal/models"
)

func mapRowToPersonnel(rows pgx.Rows) (*models.Personnel, error) {
	var p models.Personnel
	if err := rows.Scan(&p.ID, &p.Name); err != nil {
		return &models.Personnel{}, err
	}
	return &p, nil
}

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

func mapRowToScheduleType(rows pgx.Rows) (*models.ScheduleType, error) {
	var s models.ScheduleType
	if err := rows.Scan(
		&s.ID,
		&s.Name,
		&s.Description,
	); err != nil {
		return &models.ScheduleType{}, err
	}
	return &s, nil
}

func mapRowToSheduleTypePersonnel(rows pgx.Rows) (*models.ScheduleTypePersonnel, error) {
	var s models.ScheduleTypePersonnel
	if err := rows.Scan(
		&s.Personnel,
		&s.ScheduleType,
	); err != nil {
		return &models.ScheduleTypePersonnel{}, err
	}
	return &s, nil
}
