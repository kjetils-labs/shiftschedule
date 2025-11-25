package postgres

import (
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"github.com/shiftschedule/internal/models"
)

func mapRowToPersonnel(rows pgx.Rows) (*models.Personnel, error) {
	var p models.Personnel
	if err := rows.Scan(&p.ID, &p.Name); err != nil {
		return &models.Personnel{}, err
	}
	return &p, nil
}

// GetPersonnel gets current existing personnel.
func (p *Postgres) GetPersonnel() ([]*models.Personnel, error) {
	query := `
		SELECT p.id, p.name
		FROM personnel p
		GROUP BY p.id, p.name
		ORDER BY p.id;
	`

	return Query(p, query, mapRowToPersonnel)
}

// GetPersonnel gets current existing personnel.
func (p *Postgres) GetPersonnelByName(personnelName string) ([]*models.Personnel, error) {
	query := `
		SELECT p.id, p.name
		FROM personnel p
		WHERE p.name = $1
		GROUP BY p.id, p.name
		ORDER BY p.id;
	`

	return Query(p, query, mapRowToPersonnel, personnelName)
}

// GetPersonnelSchedule gets the personnel's assigned schedules.
func (p *Postgres) GetPersonnelSchedule(personnelName string) ([]*models.ShiftSchedule, error) {
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

	return Query(p, query, mapRowToShiftSchedule, personnelName)
}

// NewPersonnel creates a person.
func (p *Postgres) NewPersonnel(personnelNames []string) error {
	logger := zerolog.Ctx(p.Ctx)
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

	logger.Debug().Ctx(p.Ctx).Int("args_count", len(args)).Str("query", query).Send()

	err := p.Execute(query, args...)
	if err != nil {
		return err
	}

	return nil
}

// UpdatePersonnel creates a person.
func (p *Postgres) UpdatePersonnel(personnelName, newPersonnelName string) ([]*models.ShiftSchedule, error) {
	query := `
	UPDATE shiftschedule
	WHERE p.name = $1
	SET p.name = $2
	`
	return Query(p, query, mapRowToShiftSchedule, personnelName, newPersonnelName)
}

// UpdatePersonnel creates a person.
func (p *Postgres) DeletePersonnel(personnelName string) error {
	query := `
		DELETE FROM shiftschedule
		WHERE p.name = $1
	`

	_, err := Query(p, query, mapRowToShiftSchedule, personnelName)

	return err
}
