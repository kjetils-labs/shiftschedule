package test

import (
	"context"
	"fmt"

	"github.com/shiftschedule/internal/clients/postgres"
)

func InitTestData(ctx context.Context, pg *postgres.Postgres) error {

	if err := pg.CreateTables(); err != nil {
		return fmt.Errorf("creating tables: %w", err)
	}

	if err := pg.Execute(`
		INSERT INTO personnel (name)
		VALUES 
			('Alice'),
			('Bob'),
			('Charlie'),
			('Diana'),
			('Ethan'),
			('Fiona');
	`); err != nil {
		return fmt.Errorf("inserting personnel: %w", err)
	}

	if err := pg.Execute(`
		INSERT INTO schedule_type (name, description)
		VALUES
			('Primary TRD', 'Main Trondheim schedule'),
			('Primary OSL', 'Main Oslo schedule');
	`); err != nil {
		return fmt.Errorf("inserting schedule types: %w", err)
	}

	if err := pg.Execute(`
		INSERT INTO schedule_type_personnel (schedule_type_id, personnel_id)
		VALUES
			(1, 1), (1, 2), (1, 3),  -- Alice, Bob, Charlie eligible for Primary TRD
			(2, 4), (2, 5), (2, 6),  -- Diana, Ethan, Fiona eligible for Primary OSL
			(1, 4);                  -- Diana also eligible for Primary TRD
	`); err != nil {
		return fmt.Errorf("inserting schedule type-personnel links: %w", err)
	}

	if err := pg.Execute(`
		INSERT INTO shiftschedule (name, weeknumber, accepted, schedule_type_id)
		VALUES
			('Primary TRD - Week 1', 1, FALSE, 1),
			('Primary OSL - Week 1', 1, FALSE, 2);
	`); err != nil {
		return fmt.Errorf("inserting shift schedules: %w", err)
	}

	personnel, err := pg.GetPersonnel()
	if err != nil {
		return fmt.Errorf("fetching personnel: %w", err)
	}

	for _, person := range personnel {
		fmt.Printf("Loaded personnel: %+v\n", person)
	}

	return nil
}
