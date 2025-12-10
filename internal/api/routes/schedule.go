package routes

import (
	"net/http"

	"github.com/shiftschedule/internal/clients/postgres"
)

type ScheduleHandler struct {
	pg *postgres.Postgres
}

func (s *ScheduleHandler) GetScheduleAll(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *ScheduleHandler) GetScheduleByName(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *ScheduleHandler) NewSchedule(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *ScheduleHandler) UpdateSchedule(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *ScheduleHandler) DeleteSchedule(w http.ResponseWriter, r *http.Request) error {
	return nil
}
