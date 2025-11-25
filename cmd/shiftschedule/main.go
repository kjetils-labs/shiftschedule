package main

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/shiftschedule/internal/api"
	"github.com/shiftschedule/internal/clients/postgres"
	"github.com/shiftschedule/internal/config"
	"github.com/shiftschedule/internal/logging"
)

func main() {

	ctx, err := logging.New(context.Background(), 0)
	if err != nil {
		panic(err)
	}

	logger := zerolog.Ctx(ctx)

	config, err := config.Init()
	if err != nil {
		panic(err)
	}
	logger.Info().Msg("configuration loaded")

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

	pg, err := postgres.Init(ctx, pgConfig)
	if err != nil {
		panic(fmt.Errorf("failed to start postgres init. %w", err))
	}
	logger.Info().Msg("postgres initialization finished")

	address := fmt.Sprintf("%v:%d", config.APIHostname, config.APIPort)
	api.StartListening(ctx, pg, address)
}
