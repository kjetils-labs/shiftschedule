package postgres

import (
	"errors"

	"github.com/rs/zerolog"
)

func (dbc *DatabaseConnection) CreateTables() error {

	err := dbc.createTablePersonnel()
	if err != nil {
		return err
	}

	err = dbc.createTableShiftschedule()
	if err != nil {
		return err
	}

	err = dbc.createTableScheduleType()
	if err != nil {
		return err
	}

	err = dbc.createTableScheduleRelation()
	if err != nil {
		return err
	}

	return nil
}

func (dbc *DatabaseConnection) createTablePersonnel() error {

	query := `
		CREATE TABLE personnel (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL
		);
		`
	err := dbc.CreateTable(query, "personnel")
	if err != nil {
		return err
	}

	return nil
}

func (dbc *DatabaseConnection) createTableShiftschedule() error {

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
	err := dbc.CreateTable(query, "shiftschedule")
	if err != nil {
		return err
	}

	return nil
}
func (dbc *DatabaseConnection) createTableScheduleType() error {

	query := `
		CREATE TABLE schedule_type (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			description TEXT
		);
		`
	err := dbc.CreateTable(query, "schedule_type")
	if err != nil {
		return err
	}

	return nil
}
func (dbc *DatabaseConnection) createTableScheduleRelation() error {

	query := `
		CREATE TABLE schedule_relation (
			schedule_type_id INT REFERENCES schedule_type(id) ON DELETE CASCADE,
			personnel_id INT REFERENCES personnel(id) ON DELETE CASCADE,
			PRIMARY KEY (schedule_type_id, personnel_id)
		);
		`
	err := dbc.CreateTable(query, "schedule_relation")
	if err != nil {
		return err
	}

	return nil
}

func (dbc *DatabaseConnection) CreateTable(name string, query string) error {

	logger := zerolog.Ctx(dbc.Ctx)

	err := dbc.VerifyTable(name)
	if err == nil {
		logger.Info().Str("name", name).Msg("table already exists")
		return nil
	}

	logger.Info().Str("name", name).Msg("table missing, creating")
	err = Execute(dbc, query)
	if err != nil {
		return errors.Join(err, ErrTableCreateFailed, errors.New("failed to create table "+name))
	}
	logger.Info().Str("name", name).Msg("table created")
	return nil
}

// VerifyTable checks if the table named "name" exists.
func (dbc *DatabaseConnection) VerifyTable(name string) error {

	query := `
		SELECT EXISTS (
			SELECT 1
			FROM   information_schema.tables
			WHERE  table_schema = 'public'
			AND    table_name = $1
		)
		`
	err := Execute(dbc, query, name)

	if err != nil {
		return err
	}

	return nil
}
