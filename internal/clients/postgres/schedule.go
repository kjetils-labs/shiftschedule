package postgres

import (
	"fmt"

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

// GetSchedules gets all schedules in the db.
func (p *Postgres) GetSchedules() ([]*models.ShiftSchedule, error) {
	query := `
		SELECT 
			s.id,
			s.name,
			s.weeknumber,
			s.assignee,
			s.substitute,
			s.comment,
			s.accepted
		FROM shiftschedule s
	`
	schedules, err := Query(p, query, mapRowToShiftSchedule)
	if err != nil {
		return nil, fmt.Errorf("failed to query db. %w", err)
	}

	return schedules, nil
}

// GetSchedule get the schedule with name from the db.
func (p *Postgres) GetSchedule(name string) ([]*models.ShiftSchedule, error) {
	query := `
		SELECT 
			s.id,
			s.name,
			s.weeknumber,
			s.assignee,
			s.substitute,
			s.comment,
			s.accepted
		FROM shiftschedule s
		WHERE s.name = $1
	`
	schedules, err := Query(p, query, mapRowToShiftSchedule, name)
	if err != nil {
		return nil, fmt.Errorf("failed to query db. %w", err)
	}

	return schedules, nil
}

// GetSchedulePersonnel gets all personnel assigned a specific schedule.
func (p *Postgres) GetSchedulePersonnel(scheduleName string) (*models.ScheduleTypePersonnel, error) {
	query := `
		SELECT p.id, p.name
		FROM personnel p
		JOIN schedule_personnel sp ON p.id = sp.personnel_id
		JOIN shiftschedule s ON s.id = sp.schedule_id
		WHERE s.name = $1
		GROUP BY p.id, p.name
		ORDER BY p.id;
	`
	personnel, err := Query(p, query, mapRowToPersonnel, scheduleName)
	if err != nil {
		return nil, fmt.Errorf("failed to query db. %w", err)
	}

	result := models.ScheduleTypePersonnel{
		Personnel: personnel,
		// TODO: Fix so entire schedule is here
		ScheduleType: models.ScheduleType{Name: scheduleName},
	}
	return &result, nil
}
