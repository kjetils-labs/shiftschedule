package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/shiftschedule/internal/api/httpsuite"
	"github.com/shiftschedule/internal/domain"
)

const dateLayout = "2006-01-02"

func send[T any](r *http.Request, w http.ResponseWriter, status int, body T) {
	httpsuite.SendResponse(r.Context(), w, "", status, &body)
}

// decode reads and validates a JSON body (validate tags).
func (h *Handler) decode(w http.ResponseWriter, r *http.Request, into any) bool {
	if err := json.NewDecoder(r.Body).Decode(into); err != nil {
		h.writeError(w, r, fmt.Errorf("%w: malformed JSON body", domain.ErrInvalid))
		return false
	}
	if verr := httpsuite.IsRequestValid(into); verr != nil {
		httpsuite.SendResponse(r.Context(), w, "validation error", http.StatusBadRequest, verr)
		return false
	}
	return true
}

func parseDate(s string) (time.Time, error) {
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: date must be formatted as YYYY-MM-DD", domain.ErrInvalid)
	}
	return t, nil
}

func parseID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("%w: id must be a positive integer", domain.ErrInvalid)
	}
	return id, nil
}

// PingResult is the body of Ping.
type PingResult struct {
	Message string `json:"message" example:"pong"`
}

// RotationResult is the body of GenerateRotation.
// TODO: Move to models package?
// TODO: Is this really needed? A status code is enough to indicate success, and the number of created assignments is not that interesting.
type RotationResult struct {
	Created int `json:"created"`
}

// Ping is a public liveness check.
//
//	@Summary	Liveness check
//	@Tags		system
//	@Produce	json
//	@Success	200	{object}	httpsuite.Response[api.PingResult]
//	@Router		/ping [get]
func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	send(r, w, http.StatusOK, PingResult{Message: "pong"})
}

// Me returns the signed-in person.
//
//	@Summary	The signed-in person
//	@Tags		me
//	@Produce	json
//	@Security	BearerAuth
//	@Security	DevUser
//	@Success	200	{object}	httpsuite.Response[domain.Personnel]
//	@Failure	401	{object}	httpsuite.ErrorResponse
//	@Failure	403	{object}	httpsuite.ErrorResponse	"Signed in, but not registered as personnel"
//	@Failure	503	{object}	httpsuite.ErrorResponse	"Identity provider unavailable"
//	@Router		/me [get]
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	send(r, w, http.StatusOK, currentUser(r.Context()))
}

// MySchedule returns previous, current, next and my next week for each position I'm a candidate for. Empty if I have none.
//
//	@Summary		My schedules
//	@Description	Previous, current, next and my next week for every position of the teams I am a member of. Weeks run Monday 08:00 to Monday 08:00 (Europe/Oslo). `is_me` marks the weeks and covers I am part of. An empty list means I have no positions.
//	@Tags			me
//	@Produce		json
//	@Security		BearerAuth
//	@Security		DevUser
//	@Success		200	{object}	httpsuite.Response[[]service.PositionSchedule]
//	@Failure		401	{object}	httpsuite.ErrorResponse
//	@Failure		403	{object}	httpsuite.ErrorResponse
//	@Router			/me/schedule [get]
func (h *Handler) MySchedule(w http.ResponseWriter, r *http.Request) {
	schedules, err := h.svc.MySchedule(r.Context(), currentUser(r.Context()))
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	send(r, w, http.StatusOK, schedules)
}

// Overview returns day shifts and on-call positions with previous, current and next person. Optional ?kind=day_shift|on_call.
//
//	@Summary		Overview of all shifts
//	@Description	Day shifts (08:00-15:30) and on-call shifts (15:30-08:00) with the previous, current and next person, plus my next shift where I take part. The other kind is left empty when `kind` is given.
//	@Tags			schedules
//	@Produce		json
//	@Security		BearerAuth
//	@Security		DevUser
//	@Param			kind	query		string	false	"Only this kind of shift"	Enums(day_shift, on_call)
//	@Success		200		{object}	httpsuite.Response[service.Overview]
//	@Failure		400		{object}	httpsuite.ErrorResponse
//	@Failure		401		{object}	httpsuite.ErrorResponse
//	@Failure		403		{object}	httpsuite.ErrorResponse
//	@Router			/schedules/overview [get]
func (h *Handler) Overview(w http.ResponseWriter, r *http.Request) {
	var kind *domain.ShiftKind
	if q := r.URL.Query().Get("kind"); q != "" {
		k := domain.ShiftKind(q)
		if !k.Valid() {
			h.writeError(w, r, fmt.Errorf("%w: kind must be day_shift or on_call", domain.ErrInvalid))
			return
		}
		kind = &k
	}

	overview, err := h.svc.Overview(r.Context(), currentUser(r.Context()), kind)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	send(r, w, http.StatusOK, overview)
}

// Positions lists all positions.
//
//	@Summary	List positions
//	@Tags		schedules
//	@Produce	json
//	@Security	BearerAuth
//	@Security	DevUser
//	@Success	200	{object}	httpsuite.Response[[]domain.Position]
//	@Failure	401	{object}	httpsuite.ErrorResponse
//	@Failure	403	{object}	httpsuite.ErrorResponse
//	@Router		/positions [get]
func (h *Handler) Positions(w http.ResponseWriter, r *http.Request) {
	positions, err := h.svc.ListPositions(r.Context())
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	send(r, w, http.StatusOK, positions)
}

// TODO: Move to models package?
type setAssignmentRequest struct {
	AssigneeID *int64 `json:"assignee_id" validate:"omitempty,gt=0"`
	WeekStart  string `json:"week_start" validate:"required" example:"2026-11-16"`
	PositionID int64  `json:"position_id" validate:"required,gt=0"`
}

// SetAssignment sets, or with a null assignee_id clears, who holds a position for a week. Admins and the team's manager only.
//
//	@Summary		Set who holds a position for a week
//	@Description	A null `assignee_id` leaves the week unfilled. The assignee must be a member of the position's team. Admins and the team manager only.
//	@Tags			assignments
//	@Accept			json
//	@Security		BearerAuth
//	@Security		DevUser
//	@Param			request	body	setAssignmentRequest	true	"The assignment"
//	@Success		204
//	@Failure		400	{object}	httpsuite.ErrorResponse
//	@Failure		401	{object}	httpsuite.ErrorResponse
//	@Failure		403	{object}	httpsuite.ErrorResponse
//	@Failure		404	{object}	httpsuite.ErrorResponse
//	@Router			/assignments [put]
func (h *Handler) SetAssignment(w http.ResponseWriter, r *http.Request) {
	var in setAssignmentRequest
	if !h.decode(w, r, &in) {
		return
	}

	week, err := parseDate(in.WeekStart)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	if err := h.svc.SetAssignment(r.Context(), currentUser(r.Context()), in.PositionID, week, in.AssigneeID); err != nil {
		h.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// TODO: Move to models package?
type generateRotationRequest struct {
	Year int `json:"year" validate:"required,gte=2000,lte=2200" example:"2027"`
}

// GenerateRotation fills in a year of assignments for all teams, leaving existing weeks alone.
//
//	@Summary		Generate a year of rotations
//	@Description	Rotates each position through its team's members in alphabetical order, one week at a time. Weeks that already have an assignee are left alone.
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Security		DevUser
//	@Param			request	body		generateRotationRequest	true	"The year"
//	@Success		201		{object}	httpsuite.Response[api.RotationResult]	"Number of assignments created"
//	@Failure		400		{object}	httpsuite.ErrorResponse
//	@Failure		401		{object}	httpsuite.ErrorResponse
//	@Failure		403		{object}	httpsuite.ErrorResponse	"Admins only"
//	@Router			/admin/rotations [post]
func (h *Handler) GenerateRotation(w http.ResponseWriter, r *http.Request) {
	var in generateRotationRequest
	if !h.decode(w, r, &in) {
		return
	}

	created, err := h.svc.GenerateRotation(r.Context(), in.Year)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	send(r, w, http.StatusCreated, RotationResult{Created: created})
}
