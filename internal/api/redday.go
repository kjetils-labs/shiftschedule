package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// TODO: Move to models package?
type setRedDayRequest struct {
	Name     string `json:"name" validate:"max=200" example:"Company day"`
	IsRedDay bool   `json:"is_red_day"`
}

// SetRedDay overrides whether a date is a red day, regardless of weekends and the computed public holidays.
//
//	@Summary		Override whether a date is a red day
//	@Description	Overrides the weekends and the computed Norwegian public holidays. Red days have no day shift left to cover.
//	@Tags			admin
//	@Accept			json
//	@Security		BearerAuth
//	@Security		DevUser
//	@Param			date	path	string			true	"Date (YYYY-MM-DD)"	example(2026-12-24)
//	@Param			request	body	setRedDayRequest	true	"The override"
//	@Success		204
//	@Failure		400	{object}	httpsuite.ErrorResponse
//	@Failure		401	{object}	httpsuite.ErrorResponse
//	@Failure		403	{object}	httpsuite.ErrorResponse	"Admins only"
//	@Router			/admin/red-days/{date} [put]
func (h *Handler) SetRedDay(w http.ResponseWriter, r *http.Request) {
	day, err := parseDate(chi.URLParam(r, "date"))
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	var in setRedDayRequest
	if !h.decode(w, r, &in) {
		return
	}

	if err := h.svc.SetRedDayOverride(r.Context(), day, in.Name, in.IsRedDay); err != nil {
		h.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DeleteRedDay removes a red day override, so the computed calendar applies again.
//
//	@Summary	Remove a red day override
//	@Tags		admin
//	@Security	BearerAuth
//	@Security	DevUser
//	@Param		date	path	string	true	"Date (YYYY-MM-DD)"	example(2026-12-24)
//	@Success	204
//	@Failure	400	{object}	httpsuite.ErrorResponse
//	@Failure	401	{object}	httpsuite.ErrorResponse
//	@Failure	403	{object}	httpsuite.ErrorResponse	"Admins only"
//	@Failure	404	{object}	httpsuite.ErrorResponse
//	@Router		/admin/red-days/{date} [delete]
func (h *Handler) DeleteRedDay(w http.ResponseWriter, r *http.Request) {
	day, err := parseDate(chi.URLParam(r, "date"))
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	if err := h.svc.DeleteRedDayOverride(r.Context(), day); err != nil {
		h.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
