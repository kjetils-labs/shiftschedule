package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq"
)

var (
	ErrMissingConfig     = errors.New("missing config")
	ErrTableCreateFailed = errors.New("failed to create table")
)

func NewURL(username, password, hostname string, port int, database string, tls bool) string {
	// urlExample := "postgres://username:password@localhost:5432/database_name"
	url := fmt.Sprintf("postgres://%v:%v@%v:%d/%v", username, password, hostname, port, database)
	return url
}

type DatabaseConnection struct {
	Ctx context.Context
	DB  *sql.DB
}

func Init(ctx context.Context, url string) (*DatabaseConnection, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	connection := DatabaseConnection{
		Ctx: ctx,
		DB:  db,
	}

	err = connection.DB.Ping()
	if err != nil {
		return nil, err
	}

	return &connection, nil
}

// Query is a generic implementation where the goal was to avoid most of the boilerplating with querying.
// Query is where you expect row(s) in return, for example a SELECT statement.
func Query[T any](dbc *DatabaseConnection, query string, rowMapper func(*sql.Rows) (T, error), args ...any) ([]T, error) {
	rows, err := dbc.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []T{}
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

// Execute is where no rows are to be returned, for example INSERT and UPDATE statements.
func Execute(dbc *DatabaseConnection, query string, args ...any) error {

	tx, err := dbc.DB.BeginTx(dbc.Ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(dbc.Ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed executing: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed commiting: %w", err)
	}

	return nil
}

func SetupDB(ctx context.Context, url string) (*DatabaseConnection, error) {
	connection, err := Init(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	err = connection.CreateTables()
	if err != nil {
		return nil, fmt.Errorf("failed to create/verify necessary tables in database: %w", err)
	}

	return connection, nil
}
