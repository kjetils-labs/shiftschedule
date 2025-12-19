package postgres

import (
	"database/sql"

	"github.com/shiftschedule/internal/models"
)

func mapRowToPersonnel(rows *sql.Rows) (*models.Personnel, error) {
	var p models.Personnel
	if err := rows.Scan(&p.ID, &p.Name); err != nil {
		return &models.Personnel{}, err
	}
	return &p, nil
}

func mapRowToShiftSchedule(rows *sql.Rows) (*models.ShiftSchedule, error) {
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

func mapRowToScheduleType(rows *sql.Rows) (*models.ScheduleType, error) {
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

func mapRowToSheduleRelation(rows *sql.Rows) (*models.ScheduleRelation, error) {
	var s models.ScheduleRelation
	if err := rows.Scan(
		&s.Personnel,
		&s.ScheduleType,
	); err != nil {
		return &models.ScheduleRelation{}, err
	}
	return &s, nil
}
