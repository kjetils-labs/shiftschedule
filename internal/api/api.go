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

func setupRoutes(r *chi.Mux) error {

	r.Route("/v1", func(r chi.Router) {
		setupPersonnel(r)
		setupSchedule(r)
		setupScheduleType(r)
	})

	return nil
}

func setupPersonnel(r chi.Router) {
	personnel := routes.PersonnelHandler{}

	r.Route("/personnel", func(r chi.Router) {
		r.Get("/", wrap(personnel.GetPersonnelAll))
		r.Get("/{id}", wrap(personnel.GetPersonnelByName))
		r.Post("/", wrap(personnel.NewPersonnel))
		r.Patch("{id}", wrap(personnel.UpdatePersonnel))
		r.Delete("/{id}", wrap(personnel.DeletePersonnel))
	})
}

func setupSchedule(r chi.Router) {
	schedule := routes.ScheduleHandler{}

	r.Route("/schedule", func(r chi.Router) {
		r.Get("/", wrap(schedule.GetScheduleAll))
	})
}

func setupScheduleType(r chi.Router) {
	scheduleType := routes.ScheduleTypeHandler{}

	r.Route("/schedule", func(r chi.Router) {
		r.Get("/", wrap(scheduleType.GetScheduleTypeAll))
	})
}
