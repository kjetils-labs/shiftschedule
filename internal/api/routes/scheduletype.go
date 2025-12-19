package routes

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shiftschedule/internal/api/httpsuite"
	"github.com/shiftschedule/internal/clients/postgres"
)

type ScheduleTypeHandler struct {
	Ctx context.Context
	Dbc *postgres.DatabaseConnection
}

func (s *ScheduleTypeHandler) GetScheduleTypeAll(w http.ResponseWriter, r *http.Request) {
	schedulestypes, err := s.Dbc.GetScheduleTypes()
	if err != nil {
		httpsuite.WriteJSONError(w, "failed to get schedule types", http.StatusInternalServerError)
	}

	httpsuite.SendResponse(s.Ctx, w, "", http.StatusOK, &schedulestypes)
}

func (s *ScheduleTypeHandler) GetScheduleTypeByName(w http.ResponseWriter, r *http.Request) {
	schedules, err := s.Dbc.GetScheduleTypeByName(chi.URLParam(r, "name"))
	if err != nil {
		httpsuite.WriteJSONError(w, "failed to get schedule types", http.StatusInternalServerError)
	}
	httpsuite.SendResponse(s.Ctx, w, "", http.StatusOK, &schedules)
}

func (s *ScheduleTypeHandler) NewScheduleType(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description" binding:"optional"`
	}

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		httpsuite.SendResponse(s.Ctx, w, "decode request", http.StatusBadRequest, &err)
		return
	}

	validationErr := httpsuite.IsRequestValid(input)
	if validationErr != nil {
		httpsuite.SendResponse(s.Ctx, w, "validation error", http.StatusBadRequest, validationErr)
		return
	}

	err = s.Dbc.NewScheduleType(input.Name, input.Description)
	if err != nil {
		writeJSONError(w, "failed to create schedule type", http.StatusInternalServerError)
		return
	}

	httpsuite.SendResponse(s.Ctx, w, "schedule type created", http.StatusCreated, httpsuite.GetEmptyResponse())
}

func (s *ScheduleTypeHandler) UpdateScheduleType(w http.ResponseWriter, r *http.Request) {
}

func (s *ScheduleTypeHandler) DeleteScheduleTypeByName(w http.ResponseWriter, r *http.Request) {
	input := nameRequest{
		Name: chi.URLParam(r, "name"),
	}

	validationErr := httpsuite.IsRequestValid(input)
	if validationErr != nil {
		httpsuite.SendResponse(s.Ctx, w, "validation error", http.StatusBadRequest, validationErr)
		return
	}

	err := s.Dbc.DeleteScheduleType(input.Name)
	if err != nil {
		httpsuite.SendResponse(s.Ctx, w, "failed to update", http.StatusInternalServerError, &err)
	}

	httpsuite.SendResponse(s.Ctx, w, "personnel deleted", http.StatusAccepted, httpsuite.GetEmptyResponse())
}
