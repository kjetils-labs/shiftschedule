package api

import (
	"net/http"

	"github.com/shiftschedule/internal/service"
)

// TODO: Move to models package?
type createCoverRequest struct {
	WeekStart  string `json:"week_start" example:"2026-11-16"`
	Scope      string `json:"scope" validate:"required,oneof=rest_of_day whole_week" enums:"rest_of_day,whole_week"`
	Reason     string `json:"reason" validate:"max=500"`
	PositionID int64  `json:"position_id" validate:"required,gt=0"`
	CoveredBy  int64  `json:"covered_by" validate:"required,gt=0"`
}

// CreateCover lets the assignee hand a position over to someone else for the rest of the day or the week.
//
//	@Summary		Cover a position
//	@Description	Hand a position over to a team member for the rest of the day (only for the current week) or for the whole week. Allowed for the assignee, the team manager and admins. `week_start` (Monday, YYYY-MM-DD) is optional and defaults to the current week. Cover across teams is not allowed.
//	@Tags			covers
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Security		DevUser
//	@Param			request	body		createCoverRequest	true	"The cover"
//	@Success		201		{object}	httpsuite.Response[service.CoverView]
//	@Failure		400		{object}	httpsuite.ErrorResponse	"Invalid request, or the coverer is not in the team"
//	@Failure		401		{object}	httpsuite.ErrorResponse
//	@Failure		403		{object}	httpsuite.ErrorResponse
//	@Failure		404		{object}	httpsuite.ErrorResponse
//	@Failure		409		{object}	httpsuite.ErrorResponse	"Overlapping cover, nothing left of the shift today, or the week has passed"
//	@Router			/covers [post]
func (h *Handler) CreateCover(w http.ResponseWriter, r *http.Request) {
	var in createCoverRequest
	if !h.decode(w, r, &in) {
		return
	}

	req := service.CoverRequest{
		PositionID: in.PositionID,
		CoveredBy:  in.CoveredBy,
		Scope:      service.CoverScope(in.Scope),
		Reason:     in.Reason,
	}
	if in.WeekStart != "" {
		week, err := parseDate(in.WeekStart)
		if err != nil {
			h.writeError(w, r, err)
			return
		}
		req.WeekStart = week
	}

	me := currentUser(r.Context())
	cover, err := h.svc.CreateCover(r.Context(), me, req)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	send(r, w, http.StatusCreated, service.CoverView{
		ID: cover.ID, CoveredBy: cover.CoveredBy, StartsAt: cover.StartsAt, EndsAt: cover.EndsAt, Reason: cover.Reason,
		IsMe: cover.CoveredBy.ID == me.ID,
	})
}

// DeleteCover removes a cover.
//
//	@Summary		Remove a cover
//	@Description	Allowed for the assignee, the person covering, the team manager and admins.
//	@Tags			covers
//	@Security		BearerAuth
//	@Security		DevUser
//	@Param			id	path	int	true	"Cover id"
//	@Success		204
//	@Failure		400	{object}	httpsuite.ErrorResponse
//	@Failure		401	{object}	httpsuite.ErrorResponse
//	@Failure		403	{object}	httpsuite.ErrorResponse
//	@Failure		404	{object}	httpsuite.ErrorResponse
//	@Router			/covers/{id} [delete]
func (h *Handler) DeleteCover(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	if err := h.svc.DeleteCover(r.Context(), currentUser(r.Context()), id); err != nil {
		h.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
