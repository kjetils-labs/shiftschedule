package api

import "net/http"

// ListPersonnel lists all people.
//
//	@Summary	List personnel
//	@Tags		admin
//	@Produce	json
//	@Security	BearerAuth
//	@Security	DevUser
//	@Success	200	{object}	httpsuite.Response[[]domain.Personnel]
//	@Failure	401	{object}	httpsuite.ErrorResponse
//	@Failure	403	{object}	httpsuite.ErrorResponse	"Admins only"
//	@Router		/admin/personnel [get]
func (h *Handler) ListPersonnel(w http.ResponseWriter, r *http.Request) {
	people, err := h.svc.ListPersonnel(r.Context())
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	send(r, w, http.StatusOK, people)
}

type createPersonnelRequest struct {
	Name    string   `json:"name" validate:"required,max=200" example:"Nina Berg"`
	Email   string   `json:"email" validate:"required,email" example:"nina@example.com"`
	Groups  []string `json:"groups" validate:"dive,required,max=200" example:"alpha"`
	IsAdmin bool     `json:"is_admin"`
}

// CreatePersonnel registers a person, or updates the one with the same email.
//
//	@Summary		Register a person
//	@Description	People are matched to their login by email. Their groups decide which teams they are in.
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Security		DevUser
//	@Param			request	body		createPersonnelRequest	true	"The person"
//	@Success		201		{object}	httpsuite.Response[domain.Personnel]
//	@Failure		400		{object}	httpsuite.ErrorResponse
//	@Failure		401		{object}	httpsuite.ErrorResponse
//	@Failure		403		{object}	httpsuite.ErrorResponse	"Admins only"
//	@Router			/admin/personnel [post]
func (h *Handler) CreatePersonnel(w http.ResponseWriter, r *http.Request) {
	var in createPersonnelRequest
	if !h.decode(w, r, &in) {
		return
	}

	person, err := h.svc.CreatePersonnel(r.Context(), in.Name, in.Email, in.Groups, in.IsAdmin)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	send(r, w, http.StatusCreated, person)
}
