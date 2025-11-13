package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shiftschedule/internal/models"
)

var (
	ErrTableCreateFailed = errors.New("failed to create table")
)

type Postgres struct {
	Ctx  context.Context
	Pool *pgxpool.Pool
}

func Init(ctx context.Context, config *pgxpool.Config) (*Postgres, error) {
	// urlExample := "postgres://username:password@localhost:5432/database_name"
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		return nil, err
	}
	pg := Postgres{
		Ctx:  ctx,
		Pool: pool,
	}

	return &pg, nil
}

// Query is a generic implementation where the goal was to avoid most of the boilerplating with querying.
func Query[T any](p *Postgres, query string, rowMapper func(pgx.Rows) (T, error), args ...any) ([]T, error) {
	rows, err := p.Pool.Query(p.Ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []T
	for rows.Next() {
		item, err := rowMapper(rows)
		if err != nil {
			return results, fmt.Errorf("row scan failed: %w", err)
		}
		results = append(results, item)
	}

	if rows.Err() != nil {
		return results, fmt.Errorf("rows error: %w", rows.Err())
	}

	return results, nil
}

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
		CREATE TABLE schedule_personnel (
			schedule_id INTEGER REFERENCES shiftschedule(id) ON DELETE CASCADE,
			personnel_id INTEGER REFERENCES personnel(id) ON DELETE CASCADE,
			PRIMARY KEY (schedule_id, personnel_id)
		);
		`)

	if err != nil {
		return errors.Join(err, ErrTableCreateFailed, errors.New("failed to create schedule_personnel table"))
	}

	return nil
}

func (p *Postgres) Execute(table string) error {

	tx, err := p.Pool.Begin(p.Ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(p.Ctx)

	_, err = tx.Exec(p.Ctx, table)
	if err != nil {
		return fmt.Errorf("failed executing: %w", err)
	}
	return nil
}

// GetSchedules gets all schedules in the db.
func (p *Postgres) GetSchedules() {
}

// GetSchedulePersonnel gets all personnel assigned a specific schedule.
func (p *Postgres) GetSchedulePersonnel(scheduleName string) (*models.SchedulePersonnel, error) {
	query := `
		SELECT p.id, p.name
		FROM personnel p
		JOIN schedule_personnel sp ON p.id = sp.personnel_id
		JOIN shiftschedule s ON s.id = sp.schedule_id
		WHERE s.name = $1
		GROUP BY p.id, p.name
		ORDER BY p.id;
	`
	personnel, err := Query(p, query, mapRowToPersonnel, scheduleName)
	if err != nil {
		return nil, fmt.Errorf("failed to query db. %w", err)
	}

	result := models.SchedulePersonnel{
		Personnel:    personnel,
		ScheduleName: scheduleName,
	}
	return &result, nil
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
