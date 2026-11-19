package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/shiftschedule/internal/api/httpsuite"
	"github.com/shiftschedule/internal/auth"
	"github.com/shiftschedule/internal/domain"
)

type ctxKey struct{}

// currentUser returns the authenticated person set by authenticate.
func currentUser(ctx context.Context) domain.Personnel {
	p, _ := ctx.Value(ctxKey{}).(domain.Personnel)
	return p
}

// authenticate identifies the caller and loads their personnel record (matched on email).
func (h *Handler) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, err := h.auth.Authenticate(r)
		if err != nil {
			if errors.Is(err, auth.ErrProviderUnavailable) {
				h.writeError(w, r, err)
			} else {
				h.writeError(w, r, auth.ErrUnauthenticated)
			}
			return
		}

		person, err := h.svc.PersonByEmail(r.Context(), identity.Email)
		if errors.Is(err, domain.ErrNotFound) {
			h.writeError(w, r, domain.ErrForbidden)
			return
		}
		if err != nil {
			h.writeError(w, r, err)
			return
		}

		if identity.HasGroups {
			if person, err = h.svc.SyncGroups(r.Context(), person, identity.Groups); err != nil {
				h.writeError(w, r, err)
				return
			}
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, person)))
	})
}

func (h *Handler) requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !currentUser(r.Context()).IsAdmin {
			h.writeError(w, r, domain.ErrForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// writeError maps errors to HTTP statuses. Unknown errors are logged and hidden from the caller.
func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, err error) {
	status, msg := http.StatusInternalServerError, "internal server error"

	switch {
	case errors.Is(err, auth.ErrUnauthenticated):
		status, msg = http.StatusUnauthorized, "unauthenticated"
	case errors.Is(err, auth.ErrProviderUnavailable):
		status, msg = http.StatusServiceUnavailable, "identity provider unavailable"
		h.logger.Error().Err(err).Msg("identity provider unavailable")
	case errors.Is(err, domain.ErrNotFound):
		status, msg = http.StatusNotFound, err.Error()
	case errors.Is(err, domain.ErrForbidden):
		status, msg = http.StatusForbidden, err.Error()
	case errors.Is(err, domain.ErrInvalid), errors.Is(err, domain.ErrNotTeamMember):
		status, msg = http.StatusBadRequest, err.Error()
	case errors.Is(err, domain.ErrConflict), errors.Is(err, domain.ErrNoShiftRemaining), errors.Is(err, domain.ErrWeekAlreadyPassed):
		status, msg = http.StatusConflict, err.Error()
	default:
		h.logger.Error().Err(err).Str("path", r.URL.Path).Msg("request failed")
	}

	httpsuite.SendResponse[struct{}](r.Context(), w, msg, status, nil)
}
