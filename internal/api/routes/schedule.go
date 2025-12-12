package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shiftschedule/internal/clients/postgres"
)

type ScheduleHandler struct {
	pg *postgres.Postgres
}

func (s *ScheduleHandler) GetScheduleAll(w http.ResponseWriter, r *http.Request) error {
	schedules, err := s.pg.GetSchedules()
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(schedules)
	if err != nil {
		writeJSONError(w, "failed to encode response", http.StatusInternalServerError)
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	return nil
}

func (s *ScheduleHandler) GetScheduleByName(w http.ResponseWriter, r *http.Request) error {
	schedules, err := s.pg.GetScheduleByName(chi.URLParam(r, "name"))
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(schedules)
	if err != nil {
		writeJSONError(w, "failed to encode response", http.StatusInternalServerError)
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	return nil
}

func (s *ScheduleHandler) NewSchedule(w http.ResponseWriter, r *http.Request) error {
	var input struct {
		Name       string `json:"name" binding:"required"`
		WeekNumber int    `json:"weeknumber" binding:"required"`
		Assignee   *int   `json:"assignee" binding:"required"`
		Substitute *int   `json:"substitute" binding:"required"`
		Comment    string `json:"comment" binding:"required"`
		ScheduleID int    `json:"scheduleId" binding:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSONError(w, "failed to bind input", http.StatusBadRequest)
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	err := s.pg.NewSchedule(input.Name, input.WeekNumber, input.Assignee, input.Substitute, input.Comment, input.ScheduleID)
	if err != nil {
		writeJSONError(w, "failed to update database", http.StatusInternalServerError)
		return err
	}

	w.WriteHeader(http.StatusCreated)

	return nil
}

func (s *ScheduleHandler) UpdateSchedule(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *ScheduleHandler) DeleteSchedule(w http.ResponseWriter, r *http.Request) error {
	return nil
}
