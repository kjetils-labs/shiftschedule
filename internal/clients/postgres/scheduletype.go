package postgres

import (
	"fmt"

	"github.com/shiftschedule/internal/models"
)

// GetSchedules gets all schedules in the db.
func (p *Postgres) GetScheduleTypes() ([]*models.ScheduleType, error) {
	query := `
		SELECT 
			s.id,
			s.name,
			s.description,
		FROM schedulel_type s
	`
	schedules, err := Query(p, query, mapRowToScheduleType)
	if err != nil {
		return nil, fmt.Errorf("failed to query db. %w", err)
	}

	return schedules, nil
}

// GetSchedule get the schedule with name from the db.
func (p *Postgres) GetScheduleType(name string) ([]*models.ScheduleType, error) {
	query := `
		SELECT 
			s.id,
			s.name,
			s.description,
		FROM schedule_type s
		WHERE s.name = $1
	`
	schedules, err := Query(p, query, mapRowToScheduleType, name)
	if err != nil {
		return nil, fmt.Errorf("failed to query db. %w", err)
	}

	return schedules, nil
}

// UpdateScheduleType updates a schedule type.
func (p *Postgres) UpdateScheduleType(name string, newSchedule *models.ScheduleType) error {
	query := `
	UPDATE schedule_type s
	WHERE s.name = $1
	SET p.name = $2
	SET p.description = $3
	`
	_, err := Query(p, query, mapRowToScheduleType, name, newSchedule.Name, newSchedule.Description)
	return err
}

// DeleteScheduleType deletes a schedule type.
func (p *Postgres) DeleteScheduleType(name string) error {
	query := `
		DELETE FROM schedule_type s
		WHERE s.name = $1
	`

	_, err := Query(p, query, mapRowToShiftSchedule, name)

	return err
}
