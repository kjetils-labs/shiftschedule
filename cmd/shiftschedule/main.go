package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shiftschedule/internal/clients/postgres"
	"github.com/shiftschedule/test"
)

func main() {

	url := "postgres://user:Password123@localhost:5432/shiftschedule"
	ctx := context.Background()
	config, err := pgxpool.ParseConfig(url)
	config.ConnConfig.TLSConfig = nil
	if err != nil {
		panic(fmt.Errorf("failed to parse config from url %v. %w", url, err))
	}
	pg, err := postgres.Init(ctx, config)
	if err != nil {
		panic(fmt.Errorf("failed to start postgres init. %w", err))
	}

	err = test.InitTestData(ctx, pg)
	if err != nil {
		panic(fmt.Errorf("failed to init test data. %w", err))
	}
	// api.InitHttpServer(ctx)
}
