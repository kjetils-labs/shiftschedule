package api

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	setupRoutes(r)
	srv := &http.Server{
		Handler: r,
		Addr:    address,
	}

	return srv, nil
}

func setupRoutes(router *chi.Mux) error {

	router.Mount("/v1", setupV1Routes())
	//
	// healthcheck := v1.Group("/healthcheck")
	// healthcheck.GET("/ping", routes.Ping())
	//
	// personnel := v1.Group("/personnel")
	// personnel.GET("", routes.GetPersonnelAll(pg))
	// personnel.GET("/:personnelName", routes.GetPersonnelByName(pg))
	// personnel.POST("", routes.NewPersonnel(pg))
	// personnel.PUT("/:personelKey", routes.UpdatePersonnel(pg))
	// personnel.DELETE("/:personelKey", routes.DeletePersonnel(pg))

	return nil
}

func setupV1Routes() chi.Router {
	r := chi.NewRouter()

	r.Route("/v1", setupPersonnel)
	r.Route("/v1", setupSchedule)
	r.Route("/v1", setupScheduleType)

	return r
}

func setupPersonnel(r chi.Router) {
	personnel := routes.PersonnelHandler{}
	r.Get("/personnel", wrap(personnel.GetPersonnelAll))
}

func setupSchedule(r chi.Router) {
	schedule := routes.ScheduleHandler{}
	r.Get("/schedule", wrap(schedule.GetScheduleAll))
}

func setupScheduleType(r chi.Router) {
	scheduleType := routes.ScheduleTypeHandler{}
	r.Get("/scheduletype", wrap(scheduleType.GetScheduleTypeAll))
}
