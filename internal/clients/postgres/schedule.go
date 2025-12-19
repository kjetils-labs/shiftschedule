package postgres

import (
	"fmt"

	"github.com/shiftschedule/internal/models"
)

// GetSchedules gets all schedules in the db.
func (dbc *DatabaseConnection) GetSchedules() ([]*models.ShiftSchedule, error) {
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
	schedules, err := Query(dbc, query, mapRowToShiftSchedule)
	if err != nil {
		return nil, fmt.Errorf("failed to query db. %w", err)
	}

	return schedules, nil
}

// GetSchedule get the schedule with name from the db.
func (dbc *DatabaseConnection) GetScheduleByName(name string) ([]*models.ShiftSchedule, error) {
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
	schedules, err := Query(dbc, query, mapRowToShiftSchedule, name)
	if err != nil {
		return nil, fmt.Errorf("failed to query db. %w", err)
	}

	return schedules, nil
}

// GetSchedule get the schedule with name from the db.
func (dbc *DatabaseConnection) GetScheduleByWeek(week int) ([]*models.ShiftSchedule, error) {
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
	schedules, err := Query(dbc, query, mapRowToShiftSchedule, week)
	if err != nil {
		return nil, fmt.Errorf("failed to query db. %w", err)
	}

	return schedules, nil
}

// NewSchedule creates a new schedule.
func (dbc *DatabaseConnection) NewSchedule(name string, weeknumber int, assignee *int, substitute *int, comment string, scheduleTypeID int) error {
	query := `
		INSERT INTO shfitschedule (name, weeknumber, assignee, substitute, comment, schedule_type_id accepted)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	args := []any{name, weeknumber, assignee, substitute, comment, scheduleTypeID, false}

	err := Execute(dbc, query, args...)
	if err != nil {
		return err
	}

	return nil
}

// UpdateSchedule updates a schedule.
func (dbc *DatabaseConnection) UpdateSchedule(personnelName, newPersonnelName string) error {
	query := `
	UPDATE shiftschedule s
	WHERE s.name = $1
	SET p.name = $2
	`
	_, err := Query(dbc, query, mapRowToShiftSchedule, personnelName, newPersonnelName)

	return err
}

// DeleteSchedule deletes a schedule.
func (dbc *DatabaseConnection) DeleteSchedule(personnelName string) error {
	query := `
		DELETE FROM shiftschedule s
		WHERE s.name = $1
	`

	_, err := Query(dbc, query, mapRowToShiftSchedule, personnelName)

	return err
}

// GetSchedulePersonnel gets all personnel assigned a specific schedule.
func (dbc *DatabaseConnection) GetSchedulePersonnel(scheduleName string) ([]*models.ScheduleRelation, error) {
	query := `
		SELECT p.id, p.name
		FROM personnel p
		JOIN schedule_personnel sp ON p.id = sp.personnel_id
		JOIN shiftschedule s ON s.id = sp.schedule_id
		WHERE s.name = $1
		GROUP BY p.id, p.name
		ORDER BY p.id;
	`
	result, err := Query(dbc, query, mapRowToSheduleRelation, scheduleName)
	if err != nil {
		return nil, fmt.Errorf("failed to query db. %w", err)
	}

	return result, nil
}
