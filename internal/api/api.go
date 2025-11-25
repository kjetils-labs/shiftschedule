package api

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/shiftschedule/internal/api/routes"
	"github.com/shiftschedule/internal/clients/postgres"
)

func StartListening(ctx context.Context, pg *postgres.Postgres, address string) {

	logger := zerolog.Ctx(ctx)
	srv, err := InitHttpServer(ctx, pg, address)

	go func() {
		logger.Info().Ctx(ctx).Str("address", srv.Addr).Msg("starting api")
		err := srv.ListenAndServe()
		if err != nil {
			logger.Error().Ctx(ctx).Err(err).Send()
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	logger.Info().Ctx(ctx).Msg("shutdown signal receieved")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = srv.Shutdown(ctx)
	if err != nil {
		logger.Error().Ctx(ctx).Err(err).Msg("server forced to shutdown")
	}

	logger.Info().Ctx(ctx).Msg("server shutdown gracefully")
}

func InitHttpServer(ctx context.Context, pg *postgres.Postgres, address string) (*http.Server, error) {

	router := gin.Default()
	setupRoutes(router, pg)
	srv := &http.Server{
		Handler: router,
		Addr:    address,
	}

	return srv, nil
}

func setupRoutes(router *gin.Engine, pg *postgres.Postgres) error {
	v1 := router.Group("/v1")

	healthcheck := v1.Group("/healthcheck")
	healthcheck.GET("/ping", routes.Ping())

	personnel := v1.Group("/personnel")
	personnel.GET("", routes.GetPersonnelAll(pg))
	personnel.GET("/:personnelName", routes.GetPersonnelByName(pg))
	personnel.POST("", routes.NewPersonnel(pg))
	personnel.PUT("/:personelKey", routes.UpdatePersonnel(pg))
	personnel.DELETE("/:personelKey", routes.DeletePersonnel(pg))

	return nil
}
