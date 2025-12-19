package routes

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shiftschedule/internal/api/httpsuite"
	"github.com/shiftschedule/internal/clients/postgres"
)

type ScheduleHandler struct {
	Ctx context.Context
	Dbc *postgres.DatabaseConnection
}

func (s *ScheduleHandler) GetScheduleAll(w http.ResponseWriter, r *http.Request) {
	schedules, err := s.Dbc.GetSchedules()
	if err != nil {
		httpsuite.WriteJSONError(w, "failed to get schedules", http.StatusInternalServerError)
	}

	httpsuite.SendResponse(s.Ctx, w, "", http.StatusOK, &schedules)
}

func (s *ScheduleHandler) GetScheduleByName(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Name string `json:"name" binding:"required"`
	}

	input.Name = chi.URLParam(r, "name")

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpsuite.SendResponse(s.Ctx, w, "failed to bind input", http.StatusBadRequest, &err)
		return
	}

	validationErr := httpsuite.IsRequestValid(input)
	if validationErr != nil {
		httpsuite.SendResponse(s.Ctx, w, "validation error", http.StatusBadRequest, validationErr)
		return
	}
	schedules, err := s.Dbc.GetScheduleByName(chi.URLParam(r, "name"))
	if err != nil {
		httpsuite.WriteJSONError(w, "failed to get schedules", http.StatusInternalServerError)
	}
	httpsuite.SendResponse(s.Ctx, w, "", http.StatusOK, &schedules)
}

func (s *ScheduleHandler) NewSchedule(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name       string `json:"name" binding:"required"`
		WeekNumber int    `json:"weeknumber" binding:"required"`
		Assignee   *int   `json:"assignee" binding:"required"`
		Substitute *int   `json:"substitute" binding:"required"`
		Comment    string `json:"comment" binding:"required"`
		ScheduleID int    `json:"scheduleId" binding:"required"`
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

	err = s.Dbc.NewSchedule(input.Name, input.WeekNumber, input.Assignee, input.Substitute, input.Comment, input.ScheduleID)
	if err != nil {
		writeJSONError(w, "failed to update database", http.StatusInternalServerError)
		return
	}

	httpsuite.SendResponse(s.Ctx, w, "personnel created", http.StatusCreated, httpsuite.GetEmptyResponse())
}

func (s *ScheduleHandler) UpdateSchedule(w http.ResponseWriter, r *http.Request) {
}

func (s *ScheduleHandler) DeleteScheduleByName(w http.ResponseWriter, r *http.Request) {
	input := nameRequest{
		Name: chi.URLParam(r, "name"),
	}

	validationErr := httpsuite.IsRequestValid(input)
	if validationErr != nil {
		httpsuite.SendResponse(s.Ctx, w, "validation error", http.StatusBadRequest, validationErr)
		return
	}

	err := s.Dbc.DeleteSchedule(input.Name)
	if err != nil {
		httpsuite.SendResponse(s.Ctx, w, "failed to update", http.StatusInternalServerError, &err)
	}

	httpsuite.SendResponse(s.Ctx, w, "schedule deleted", http.StatusAccepted, httpsuite.GetEmptyResponse())
}
