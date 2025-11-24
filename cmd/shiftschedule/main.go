package main

import (
	"context"
	"fmt"

	"github.com/shiftschedule/internal/clients/postgres"
	"github.com/shiftschedule/internal/config"
	"github.com/shiftschedule/test"
)

func main() {

	enableTLS := false
	config, err := config.Init()
	if err != nil {
		panic(err)
	}

	pgConfig, err := postgres.NewPostgresConfig(
		config.PostgresUsername,
		config.PostgresPassword,
		config.PostgresHostname,
		config.PostgresPort,
		config.PostgresDatabase,
		enableTLS,
	)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	pg, err := postgres.Init(ctx, pgConfig)
	if err != nil {
		panic(fmt.Errorf("failed to start postgres init. %w", err))
	}

	err = test.InitTestData(ctx, pg)
	if err != nil {
		panic(fmt.Errorf("failed to init test data. %w", err))
	}
	// api.InitHttpServer(ctx)
}
