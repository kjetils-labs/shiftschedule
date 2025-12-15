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
		r.Get("/", personnel.GetPersonnelAll)
		r.Get("/{name}", personnel.GetPersonnelByName)
		r.Post("/", personnel.NewPersonnel)
		r.Patch("{id}", personnel.UpdatePersonnel)
		r.Delete("/{id}", personnel.DeletePersonnel)
	})
}

func setupSchedule(r chi.Router) {
	schedule := routes.ScheduleHandler{}

	r.Route("/schedule", func(r chi.Router) {
		r.Get("/", wrap(schedule.GetScheduleAll))
		r.Get("/{name}", wrap(schedule.GetScheduleByName))
		r.Post("/", wrap(schedule.NewSchedule))
		r.Patch("/{id}", wrap(schedule.UpdateSchedule))
		r.Delete("/{id}", wrap(schedule.DeleteSchedule))
	})
}

func setupScheduleType(r chi.Router) {
	scheduleType := routes.ScheduleTypeHandler{}

	r.Route("/scheduletype", func(r chi.Router) {
		r.Get("/", wrap(scheduleType.GetScheduleTypeAll))
		r.Get("/{name}", wrap(scheduleType.GetScheduleTypeByName))
		r.Post("/", wrap(scheduleType.NewScheduleType))
		r.Patch("/{id}", wrap(scheduleType.UpdateScheduleType))
		r.Delete("/{id}", wrap(scheduleType.DeleteScheduleType))
	})
}
