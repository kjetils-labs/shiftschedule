// Package api sets up the HTTP server and routes.
package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	_ "github.com/shiftschedule/internal/api/swagger" // registers the generated spec
	"github.com/shiftschedule/internal/auth"
	"github.com/shiftschedule/internal/service"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

const (
	shutdownTimeout   = 5 * time.Second
	readHeaderTimeout = 5 * time.Second
)

//go:generate swag init -g internal/api/api.go -d ../.. --exclude node_modules,web,hacks,docs -o swagger --parseInternal --parseDependency --useStructName --outputTypes go,json,yaml

//	@title			Shift schedule API
//	@version		1.0
//	@description	Who has which shift. A week runs Monday 08:00 to Monday 08:00 (Europe/Oslo): day shifts are 08:00-15:30 and on-call 15:30-08:00 the day after.
//	@BasePath		/v1
//
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				AUTH_MODE=oidc: "Bearer <ID token from Dex>".
//
//	@securityDefinitions.apikey	DevUser
//	@in							header
//	@name						X-Dev-User
//	@description				AUTH_MODE=dev only: the email of the person to act as.

// Handler holds the dependencies of the HTTP handlers.
type Handler struct {
	svc    *service.Service
	auth   auth.Authenticator
	logger zerolog.Logger
}

// NewHandler creates the HTTP handlers.
func NewHandler(logger zerolog.Logger, svc *service.Service, authenticator auth.Authenticator) *Handler {
	return &Handler{svc: svc, auth: authenticator, logger: logger}
}

// Routes returns the router with all routes under /v1.
func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
	})
	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))

	r.Route("/v1", func(r chi.Router) {
		r.Get("/ping", h.Ping)

		r.Group(func(r chi.Router) {
			r.Use(h.authenticate)

			r.Get("/me", h.Me)
			r.Get("/me/schedule", h.MySchedule)
			r.Get("/schedules/overview", h.Overview)
			r.Get("/positions", h.Positions)
			r.Post("/covers", h.CreateCover)
			r.Delete("/covers/{id}", h.DeleteCover)
			r.Put("/assignments", h.SetAssignment)

			r.Route("/admin", func(r chi.Router) {
				r.Use(h.requireAdmin)

				r.Get("/personnel", h.ListPersonnel)
				r.Post("/personnel", h.CreatePersonnel)
				r.Post("/rotations", h.GenerateRotation)
				r.Put("/red-days/{date}", h.SetRedDay)
				r.Delete("/red-days/{date}", h.DeleteRedDay)
			})
		})
	})

	return r
}

// Serve listens on address until SIGINT/SIGTERM, then shuts down gracefully.
func Serve(ctx context.Context, logger zerolog.Logger, handler http.Handler, address string) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{Addr: address, Handler: handler, ReadHeaderTimeout: readHeaderTimeout}

	errCh := make(chan error, 1)
	go func() {
		logger.Info().Str("address", srv.Addr).Msg("starting api")
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("server stopped: %w", err)
	case <-ctx.Done():
		logger.Info().Msg("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	logger.Info().Msg("server shutdown gracefully")
	return nil
}
