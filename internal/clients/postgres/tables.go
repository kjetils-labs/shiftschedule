package postgres

import (
	"errors"

	"github.com/rs/zerolog"
)

func (p *Postgres) CreateTables() error {

	err := p.createTablePersonnel()
	if err != nil {
		return err
	}

	err = p.createTableShiftschedule()
	if err != nil {
		return err
	}

	err = p.createTableScheduleType()
	if err != nil {
		return err
	}

	err = p.createTableScheduleTypePersonnel()
	if err != nil {
		return err
	}

	return nil
}

func (p *Postgres) createTablePersonnel() error {

	query := `
		CREATE TABLE personnel (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL
		);
		`
	err := p.CreateTable(query, "personnel")
	if err != nil {
		return err
	}

	return nil
}

func (p *Postgres) createTableShiftschedule() error {

	query := `
		CREATE TABLE shiftschedule (
			id SERIAL PRIMARY KEY,
			schedule_type_id INT REFERENCES schedule_type(id)
			name TEXT NOT NULL,
			weeknumber INTEGER NOT NULL,
			assignee INTEGER REFERENCES personnel(id),
			substitute INTEGER REFERENCES personnel(id),
			comment TEXT,
			accepted BOOLEAN DEFAULT FALSE
		);
		`
	err := p.CreateTable(query, "shiftschedule")
	if err != nil {
		return err
	}

	return nil
}
func (p *Postgres) createTableScheduleType() error {

	query := `
		CREATE TABLE schedule_type (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			description TEXT
		);
		`
	err := p.CreateTable(query, "schedule_type")
	if err != nil {
		return err
	}

	return nil
}
func (p *Postgres) createTableScheduleTypePersonnel() error {

	query := `
		CREATE TABLE schedule_type_personnel (
			schedule_type_id INT REFERENCES schedule_type(id) ON DELETE CASCADE,
			personnel_id INT REFERENCES personnel(id) ON DELETE CASCADE,
			PRIMARY KEY (schedule_type_id, personnel_id)
		);
		`
	err := p.CreateTable(query, "schedule_type_personnel")
	if err != nil {
		return err
	}

	return nil
}

func (p *Postgres) CreateTable(name string, query string) error {

	logger := zerolog.Ctx(p.Ctx)

	err := p.TableExists(name)
	if err == nil {
		logger.Info().Str("name", name).Msg("table already exists")
		return nil
	}

	logger.Info().Str("name", name).Msg("table missing, creating")
	err = p.Execute(query)
	if err != nil {
		return errors.Join(err, ErrTableCreateFailed, errors.New("failed to create table "+name))
	}
	logger.Info().Str("name", name).Msg("table created")
	return nil
}

func (p *Postgres) TableExists(name string) error {

	err := p.Execute(`
		SELECT EXISTS (
            SELECT 1
            FROM   information_schema.tables
            WHERE  table_schema = 'public'
            AND    table_name = $1
		)
		`, name)

	if err != nil {
		return err
	}

	return nil
}
