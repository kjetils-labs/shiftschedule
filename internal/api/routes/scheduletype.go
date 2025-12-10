package routes

import (
	"net/http"

	"github.com/shiftschedule/internal/clients/postgres"
)

type ScheduleTypeHandler struct {
	pg *postgres.Postgres
}

func (s *ScheduleTypeHandler) GetScheduleTypeAll(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *ScheduleTypeHandler) GetScheduleTypeByName(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *ScheduleTypeHandler) NewScheduleType(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *ScheduleTypeHandler) UpdateScheduleType(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *ScheduleTypeHandler) DeleteScheduleType(w http.ResponseWriter, r *http.Request) error {
	return nil
}
