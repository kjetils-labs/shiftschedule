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
		FROM schedulel_type s
		WHERE s.name = $1
	`
	schedules, err := Query(p, query, mapRowToScheduleType, name)
	if err != nil {
		return nil, fmt.Errorf("failed to query db. %w", err)
	}

	return schedules, nil
}
