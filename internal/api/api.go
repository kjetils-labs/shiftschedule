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

func StartListening(ctx context.Context, dbc *postgres.DatabaseConnection, address string) {

	logger := zerolog.Ctx(ctx)
	srv, err := InitHttpServer(ctx, dbc, address)

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

func InitHttpServer(ctx context.Context, dbc *postgres.DatabaseConnection, address string) (*http.Server, error) {

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	setupRoutes(ctx, r, dbc)
	srv := &http.Server{
		Handler: r,
		Addr:    address,
	}

	return srv, nil
}

func setupRoutes(ctx context.Context, r *chi.Mux, dbc *postgres.DatabaseConnection) error {

	r.Route("/v1", func(r chi.Router) {
		setupHealthcheck(r)
		setupPersonnel(ctx, dbc, r)
		setupSchedule(ctx, dbc, r)
		setupScheduleType(ctx, dbc, r)
	})

	return nil
}

func setupPersonnel(ctx context.Context, dbc *postgres.DatabaseConnection, r chi.Router) {
	handler := routes.PersonnelHandler{Ctx: ctx, Dbc: dbc}

	r.Route("/personnel", func(r chi.Router) {
		r.Get("/", handler.GetPersonnelAll)
		r.Get("/{name}", handler.GetPersonnelByName)
		r.Post("/", handler.NewPersonnel)
		r.Patch("{id}", handler.UpdatePersonnel)
		r.Delete("/{id}", handler.DeletePersonnelByName)
	})
}

func setupSchedule(ctx context.Context, dbc *postgres.DatabaseConnection, r chi.Router) {
	handler := routes.ScheduleHandler{Ctx: ctx, Dbc: dbc}

	r.Route("/schedule", func(r chi.Router) {
		r.Get("/", handler.GetScheduleAll)
		r.Get("/{name}", handler.GetScheduleByName)
		r.Post("/", handler.NewSchedule)
		r.Patch("/{id}", handler.UpdateSchedule)
		r.Delete("/{id}", handler.DeleteScheduleByName)
	})
}

func setupScheduleType(ctx context.Context, dbc *postgres.DatabaseConnection, r chi.Router) {
	handler := routes.ScheduleTypeHandler{Ctx: ctx, Dbc: dbc}

	r.Route("/scheduletype", func(r chi.Router) {
		r.Get("/", handler.GetScheduleTypeAll)
		r.Get("/{name}", handler.GetScheduleTypeByName)
		r.Post("/", handler.NewScheduleType)
		r.Patch("/{id}", handler.UpdateScheduleType)
		r.Delete("/{id}", handler.DeleteScheduleTypeByName)
	})
}

func setupHealthcheck(r chi.Router) {
	handler := routes.HealthCheckHandler{}

	r.Route("/scheduletype", func(r chi.Router) {
		r.Get("/ping", handler.Ping)
	})
}
