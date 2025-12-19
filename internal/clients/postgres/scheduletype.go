package postgres

import (
	"fmt"

	"github.com/shiftschedule/internal/models"
)

// GetSchedules gets all schedules in the db.
func (dbc *DatabaseConnection) GetScheduleTypes() ([]*models.ScheduleType, error) {
	query := `
		SELECT 
			s.id,
			s.name,
			s.description,
		FROM schedulel_type s
	`
	schedules, err := Query(dbc, query, mapRowToScheduleType)
	if err != nil {
		return nil, fmt.Errorf("failed to query db. %w", err)
	}

	return schedules, nil
}

// GetSchedule get the schedule with name from the db.
func (dbc *DatabaseConnection) GetScheduleTypeByName(name string) ([]*models.ScheduleType, error) {
	query := `
		SELECT 
			s.id,
			s.name,
			s.description,
		FROM schedule_type s
		WHERE s.name = $1
	`
	schedules, err := Query(dbc, query, mapRowToScheduleType, name)
	if err != nil {
		return nil, fmt.Errorf("failed to query db. %w", err)
	}

	return schedules, nil
}

// NewSchedule creates a new schedule.
func (dbc *DatabaseConnection) NewScheduleType(name, description string) error {
	query := `
		INSERT INTO schedule_type (name, description)
		VALUES ($1, $2)
	`

	args := []any{name, description}

	err := Execute(dbc, query, args...)
	if err != nil {
		return err
	}

	return nil
}

// UpdateScheduleType updates a schedule type.
func (dbc *DatabaseConnection) UpdateScheduleType(name string, newSchedule *models.ScheduleType) error {
	query := `
	UPDATE schedule_type s
	WHERE s.name = $1
	SET p.name = $2
	SET p.description = $3
	`
	_, err := Query(dbc, query, mapRowToScheduleType, name, newSchedule.Name, newSchedule.Description)
	return err
}

// DeleteScheduleType deletes a schedule type.
func (dbc *DatabaseConnection) DeleteScheduleType(name string) error {
	query := `
		DELETE FROM schedule_type s
		WHERE s.name = $1
	`

	_, err := Query(dbc, query, mapRowToShiftSchedule, name)

	return err
}
