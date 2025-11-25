package main_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/shiftschedule/internal/clients/postgres"
	"github.com/shiftschedule/test"
)

func TestMain(m *testing.M) {

	config := test.InitTestConfig()

	pgConfig, err := postgres.NewPostgresConfig(
		config.PostgresUsername,
		config.PostgresPassword,
		config.PostgresHostname,
		config.PostgresPort,
		config.PostgresDatabase,
		config.PostgresEnableTLS,
	)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	pg, err := postgres.Init(ctx, pgConfig)

	err = test.InitTestData(ctx, pg)
	if err != nil {
		panic(fmt.Errorf("failed to init test data. %w", err))
	}
	// api.InitHttpServer(ctx)
}
