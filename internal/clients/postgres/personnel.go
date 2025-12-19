package postgres

import (
	"fmt"

	"github.com/rs/zerolog"
	"github.com/shiftschedule/internal/models"
)

// GetPersonnel gets current existing personnel.
func (dbc *DatabaseConnection) GetPersonnel() ([]*models.Personnel, error) {
	query := `
		SELECT p.id, p.name
		FROM personnel p
		GROUP BY p.id, p.name
		ORDER BY p.id;
	`

	return Query(dbc, query, mapRowToPersonnel)
}

// GetPersonnel gets current existing personnel.
func (dbc *DatabaseConnection) GetPersonnelByName(personnelName string) ([]*models.Personnel, error) {
	query := `
		SELECT p.id, p.name
		FROM personnel p
		WHERE p.name = $1
		GROUP BY p.id, p.name
		ORDER BY p.id;
	`

	return Query(dbc, query, mapRowToPersonnel, personnelName)
}

// GetPersonnelSchedule gets the personnel's assigned schedules.
func (dbc *DatabaseConnection) GetPersonnelSchedule(personnelName string) ([]*models.ShiftSchedule, error) {
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
		JOIN schedule_personnel sp ON s.id = sp.schedule_id
		JOIN personnel p ON sp.personnel_id = p.id
		WHERE p.name = $1
		ORDER BY s.weeknumber;
	`

	return Query(dbc, query, mapRowToShiftSchedule, personnelName)
}

// NewPersonnel creates a person.
func (dbc *DatabaseConnection) NewPersonnel(personnelNames []string) error {
	for i, name := range personnelNames {
		if name == "" {
			return fmt.Errorf("name at index %d is empty", i)
		}
	}
	logger := zerolog.Ctx(dbc.Ctx)
	query := `
		INSERT INTO personnel (name)
		VALUES 
	`
	args := make([]any, 0, len(personnelNames))
	for i, name := range personnelNames {
		placeholder := fmt.Sprintf("$%d", i+1)
		if i > 0 {
			query += ", "
		}
		query += "(" + placeholder + ")"
		args = append(args, name)
	}

	logger.Debug().Ctx(dbc.Ctx).Int("args_count", len(args)).Str("query", query).Send()

	err := Execute(dbc, query, args...)
	if err != nil {
		return err
	}

	return nil
}

// UpdatePersonnel updates a person.
func (dbc *DatabaseConnection) UpdatePersonnel(personnelName, newPersonnelName string) ([]*models.ShiftSchedule, error) {
	query := `
	UPDATE personnel p
	WHERE p.name = $1
	SET p.name = $2
	`
	return Query(dbc, query, mapRowToShiftSchedule, personnelName, newPersonnelName)
}

// DeletePersonnel deletes a person.
func (dbc *DatabaseConnection) DeletePersonnel(personnelName string) error {
	query := `
		DELETE FROM personnel
		WHERE p.name = $1
	`

	_, err := Query(dbc, query, mapRowToShiftSchedule, personnelName)

	return err
}
