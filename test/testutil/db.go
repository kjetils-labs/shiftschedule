// Package testutil provides helpers for tests that need a real Postgres.
package testutil

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/shiftschedule/internal/clients/postgres"
)

const (
	envURL        = "TEST_DATABASE_URL"
	timeout       = 30 * time.Second
	randomIDBytes = 6
)

// NewStore creates a fresh, migrated database for the test and drops it afterwards.
// TEST_DATABASE_URL must point at a Postgres the user may create databases in; the test is skipped when it is unset.
func NewStore(t *testing.T) *postgres.Store {
	t.Helper()

	adminURL := lookupEnv(t)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	admin, err := pgx.Connect(ctx, adminURL)
	if err != nil {
		t.Fatalf("failed to connect to %s: %v", envURL, err)
	}
	defer admin.Close(ctx)

	buf := make([]byte, randomIDBytes)
	if _, err := rand.Read(buf); err != nil {
		t.Fatal(err)
	}
	name := "shiftschedule_test_" + hex.EncodeToString(buf)

	if _, err := admin.Exec(ctx, "CREATE DATABASE "+name); err != nil {
		t.Fatalf("failed to create database: %v", err)
	}

	u, err := url.Parse(adminURL)
	if err != nil {
		t.Fatal(err)
	}
	u.Path = "/" + name

	store, err := postgres.New(ctx, u.String())
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}
	if err := store.Migrate(ctx); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	t.Cleanup(func() {
		store.Close()

		cleanupCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		conn, err := pgx.Connect(cleanupCtx, adminURL)
		if err != nil {
			t.Logf("failed to connect for cleanup: %v", err)
			return
		}
		defer conn.Close(cleanupCtx)
		if _, err := conn.Exec(cleanupCtx, "DROP DATABASE "+name+" WITH (FORCE)"); err != nil {
			t.Logf("failed to drop %s: %v", name, err)
		}
	})

	return store
}

func lookupEnv(t *testing.T) string {
	t.Helper()
	v := os.Getenv(envURL)
	if v == "" {
		t.Skipf("%s is not set, skipping database test", envURL)
	}
	return v
}
