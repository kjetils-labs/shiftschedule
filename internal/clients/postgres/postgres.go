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

func NewPostgresConfig(username, password, hostname string, port int, database string, tls bool) (*pgxpool.Config, error) {

	// urlExample := "postgres://username:password@localhost:5432/database_name"
	url := fmt.Sprintf("postgres://%v:%v@%v:%d/%v", username, password, hostname, port, database)
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config from url %v. %w", url, err)
	}
	if !tls {
		config.ConnConfig.TLSConfig = nil
	}
	return config, nil
}

type Postgres struct {
	Ctx  context.Context
	Pool *pgxpool.Pool
}

func Init(ctx context.Context, config *pgxpool.Config) (*Postgres, error) {
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
