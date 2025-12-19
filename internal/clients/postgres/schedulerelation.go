package postgres

import (
	"fmt"

	"github.com/shiftschedule/internal/models"
)

// GetSchedules gets all schedules in the db.
func (dbc *DatabaseConnection) GetScheduleRelations() ([]*models.ScheduleRelation, error) {
	query := `
		SELECT 
			s.id,
			s.name,
			s.description,
		FROM schedulel_type s
	`
	schedulerelations, err := Query(dbc, query, mapRowToSheduleRelation)
	if err != nil {
		return nil, fmt.Errorf("failed to query db. %w", err)
	}

	return schedulerelations, nil
}

// GetScheduleRelationByPersonnelID get the schedule with name from the db.
func (dbc *DatabaseConnection) GetScheduleRelationByPersonnelID(id int) ([]*models.ScheduleType, error) {
	query := `
		SELECT 
			s.id,
			s.name,
			s.description,
		FROM schedule_type s
		WHERE s.name = $1
	`
	schedules, err := Query(dbc, query, mapRowToScheduleType, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query db. %w", err)
	}

	return schedules, nil
}
