package test

import (
	"context"
	"fmt"

	"github.com/shiftschedule/internal/clients/postgres"
)

func InitTestData(ctx context.Context, pg *postgres.Postgres) error {

	err := pg.CreateTables()
	if err != nil {
		return fmt.Errorf("creating table: %w", err)
	}

	err = pg.Execute(`
		INSERT INTO personnel (name)
		VALUES ('Alice'), ('Bob'), ('Charlie'), ('Diana'), ('Ethan'), ('Fiona');
    `)
	if err != nil {
		return fmt.Errorf("inserting data: %w", err)
	}

	personnel, err := pg.GetPersonnel()
	if err != nil {
		return fmt.Errorf("failed to get personne: %w", err)
	}

	for _, person := range personnel {
		if person == nil {
			return fmt.Errorf("person is nil")
		}
		fmt.Println(person)
	}

	err = pg.Execute(`
		INSERT INTO shiftschedule (name, weeknumber, accepted)
		VALUES ('Primary TRD', 1, FALSE),
			   ('Primary OSL', 1, FALSE);
    `)
	if err != nil {
		return fmt.Errorf("inserting data: %w", err)
	}

	err = pg.Execute(`
		INSERT INTO schedule_personnel (schedule_id, personnel_id)
		VALUES 
		  (1, 1), (1, 2), (1, 3),  -- Alice, Bob, Charlie on Primary TRD
		  (2, 4), (2, 5), (2, 6),  -- Diana, Ethan, Fiona on Backup OSL
		  (1, 4);                  -- Diana is also allowed on Primary TRD
    `)
	if err != nil {
		return fmt.Errorf("inserting data: %w", err)
	}

	return nil
}
