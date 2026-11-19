// Package postgres is the pgx based persistence layer. Schema changes live in migrations/ and are applied with goose.
package postgres

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/url"
	"strconv"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/shiftschedule/internal/domain"
)

//go:embed migrations/*.sql
var migrations embed.FS

const (
	pgExclusionViolation  = "23P01"
	pgUniqueViolation     = "23505"
	pgForeignKeyViolation = "23503"
)

// NewURL builds a connection URL, escaping credentials.
func NewURL(username, password, hostname string, port int, database string, tls bool) string {
	sslmode := "disable"
	if tls {
		sslmode = "require"
	}

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(username, password),
		Host:     net.JoinHostPort(hostname, strconv.Itoa(port)),
		Path:     database,
		RawQuery: url.Values{"sslmode": {sslmode}}.Encode(),
	}
	return u.String()
}

// Store gives access to the database.
type Store struct {
	pool *pgxpool.Pool
}

// New connects to the database and verifies the connection.
func New(ctx context.Context, connURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, connURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Store{pool: pool}, nil
}

// NewFromPool wraps an existing pool, for example in tests.
func NewFromPool(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// Close releases all connections.
func (s *Store) Close() {
	s.pool.Close()
}

// Migrate applies all pending migrations.
func (s *Store) Migrate(ctx context.Context) error {
	db := stdlib.OpenDBFromPool(s.pool)
	defer db.Close()

	sub, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("failed to open embedded migrations: %w", err)
	}

	provider, err := goose.NewProvider(goose.DialectPostgres, db, sub)
	if err != nil {
		return fmt.Errorf("failed to create migration provider: %w", err)
	}

	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}
	return nil
}

// mapError turns database errors into domain errors where the caller can act on them.
func mapError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}

	switch pgErr.Code {
	case pgExclusionViolation, pgUniqueViolation:
		return fmt.Errorf("%w: %s", domain.ErrConflict, pgErr.Message)
	case pgForeignKeyViolation:
		return fmt.Errorf("%w: %s", domain.ErrInvalid, pgErr.Message)
	default:
		return err
	}
}
