package routes

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shiftschedule/internal/api/httpsuite"
	"github.com/shiftschedule/internal/clients/postgres"
)

type PersonnelHandler struct {
	Ctx context.Context
	Dbc *postgres.DatabaseConnection
}

type nameRequest struct {
	Name string `json:"name" binding:"required"`
}

func (p *PersonnelHandler) GetPersonnelAll(w http.ResponseWriter, r *http.Request) {
	personnel, err := p.Dbc.GetPersonnel()
	if err != nil {
		httpsuite.WriteJSONError(w, "failed to get personnel", http.StatusInternalServerError)
	}

	httpsuite.SendResponse(p.Ctx, w, "", http.StatusOK, &personnel)
}

func (p *PersonnelHandler) GetPersonnelByName(w http.ResponseWriter, r *http.Request) {

	input := nameRequest{
		Name: chi.URLParam(r, "name"),
	}

	validationErr := httpsuite.IsRequestValid(input)
	if validationErr != nil {
		httpsuite.SendResponse(p.Ctx, w, "validation error", http.StatusBadRequest, &input)
		return
	}
	personnel, err := p.Dbc.GetPersonnelByName(chi.URLParam(r, "name"))
	if err != nil {
		httpsuite.WriteJSONError(w, "failed to get personnel", http.StatusInternalServerError)
	}

	httpsuite.SendResponse(p.Ctx, w, "", http.StatusOK, &personnel)
}

func (p *PersonnelHandler) NewPersonnel(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Names []string `json:"names" binding:"required,unique"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpsuite.SendResponse(p.Ctx, w, "failed to bind input", http.StatusBadRequest, &err)
		return
	}

	validationErr := httpsuite.IsRequestValid(input)
	if validationErr != nil {
		httpsuite.SendResponse(p.Ctx, w, "validation error", http.StatusBadRequest, validationErr)
		return
	}

	err := p.Dbc.NewPersonnel(input.Names)
	if err != nil {
		httpsuite.SendResponse(p.Ctx, w, "failed to update", http.StatusInternalServerError, &err)
		return
	}

	httpsuite.SendResponse(p.Ctx, w, "personnel created", http.StatusCreated, httpsuite.GetEmptyResponse())
}

func (p *PersonnelHandler) UpdatePersonnel(w http.ResponseWriter, r *http.Request) {
}

func (p *PersonnelHandler) DeletePersonnelByName(w http.ResponseWriter, r *http.Request) {
	input := nameRequest{
		Name: chi.URLParam(r, "name"),
	}

	validationErr := httpsuite.IsRequestValid(input)
	if validationErr != nil {
		httpsuite.SendResponse(p.Ctx, w, "validation error", http.StatusBadRequest, validationErr)
		return
	}

	err := p.Dbc.DeletePersonnel(input.Name)
	if err != nil {
		httpsuite.SendResponse(p.Ctx, w, "failed to update", http.StatusInternalServerError, &err)
	}

	httpsuite.SendResponse(p.Ctx, w, "personnel deleted", http.StatusAccepted, httpsuite.GetEmptyResponse())
}
