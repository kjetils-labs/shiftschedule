package postgres

import "errors"

func (p *Postgres) CreateTables() error {
	err := p.Execute(`
		CREATE TABLE personnel (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL
		);
		`)

	if err != nil {
		return errors.Join(err, ErrTableCreateFailed, errors.New("failed to create personnel table"))
	}

	err = p.Execute(`
		CREATE TABLE shiftschedule (
			id SERIAL PRIMARY KEY,
			schedule_type_id INT REFERENCES schedule_type(id);
			name TEXT NOT NULL,
			weeknumber INTEGER NOT NULL,
			assignee INTEGER REFERENCES personnel(id),
			substitute INTEGER REFERENCES personnel(id),
			comment TEXT,
			accepted BOOLEAN DEFAULT FALSE
		);
		`)

	if err != nil {
		return errors.Join(err, ErrTableCreateFailed, errors.New("failed to create shfitschedule table"))
	}

	err = p.Execute(`
		CREATE TABLE schedule_type (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			description TEXT
		);
		`)

	if err != nil {
		return errors.Join(err, ErrTableCreateFailed, errors.New("failed to create schedule_type table"))
	}

	err = p.Execute(`
		CREATE TABLE schedule_type_personnel (
			schedule_type_id INT REFERENCES schedule_type(id) ON DELETE CASCADE,
			personnel_id INT REFERENCES personnel(id) ON DELETE CASCADE,
			PRIMARY KEY (schedule_type_id, personnel_id)
		);
		`)

	if err != nil {
		return errors.Join(err, ErrTableCreateFailed, errors.New("failed to create schedule_type_personnel table"))
	}

	return nil
}
