package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrMissingConfig     = errors.New("missing config")
	ErrTableCreateFailed = errors.New("failed to create table")
)

type Postgres struct {
	Ctx  context.Context
	Pool *pgxpool.Pool
}

func Init(ctx context.Context, config *pgxpool.Config) (*Postgres, error) {
	// urlExample := "postgres://username:password@localhost:5432/database_name"
	if config == nil {
		return nil, ErrMissingConfig
	}

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

	err = tx.Commit(p.Ctx)
	if err != nil {
		return fmt.Errorf("failed commiting: %w", err)
	}

	return nil
}
