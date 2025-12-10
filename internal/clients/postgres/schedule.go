package postgres

import (
	"fmt"

	"github.com/shiftschedule/internal/models"
)

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
func (p *Postgres) GetScheduleByName(name string) ([]*models.ShiftSchedule, error) {
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

// GetSchedule get the schedule with name from the db.
func (p *Postgres) GetScheduleByWeek(week int) ([]*models.ShiftSchedule, error) {
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
		WHERE s.weeknumber = $1
	`
	schedules, err := Query(p, query, mapRowToShiftSchedule, week)
	if err != nil {
		return nil, fmt.Errorf("failed to query db. %w", err)
	}

	return schedules, nil
}

// NewSchedule creates a new schedule.
func (p *Postgres) NewSchedule(name string, weeknumber int, assignee *int, substitute *int, comment string, scheduleTypeID int) error {
	query := `
		INSERT INTO shfitschedule (name, weeknumber, assignee, substitute, comment, schedule_type_id accepted)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	args := []any{name, weeknumber, assignee, substitute, comment, scheduleTypeID, false}

	err := p.Execute(query, args...)
	if err != nil {
		return err
	}

	return nil
}

// UpdateSchedule updates a schedule.
func (p *Postgres) UpdateSchedule(personnelName, newPersonnelName string) error {
	query := `
	UPDATE shiftschedule s
	WHERE s.name = $1
	SET p.name = $2
	`
	_, err := Query(p, query, mapRowToShiftSchedule, personnelName, newPersonnelName)

	return err
}

// DeleteSchedule deletes a schedule.
func (p *Postgres) DeleteSchedule(personnelName string) error {
	query := `
		DELETE FROM shiftschedule s
		WHERE s.name = $1
	`

	_, err := Query(p, query, mapRowToShiftSchedule, personnelName)

	return err
}

// GetSchedulePersonnel gets all personnel assigned a specific schedule.
func (p *Postgres) GetSchedulePersonnel(scheduleName string) ([]*models.ScheduleTypePersonnel, error) {
	query := `
		SELECT p.id, p.name
		FROM personnel p
		JOIN schedule_personnel sp ON p.id = sp.personnel_id
		JOIN shiftschedule s ON s.id = sp.schedule_id
		WHERE s.name = $1
		GROUP BY p.id, p.name
		ORDER BY p.id;
	`
	result, err := Query(p, query, mapRowToSheduleTypePersonnel, scheduleName)
	if err != nil {
		return nil, fmt.Errorf("failed to query db. %w", err)
	}

	return result, nil
}
