package routes

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shiftschedule/internal/api/httpsuite"
	"github.com/shiftschedule/internal/clients/postgres"
)

type ScheduleRelationHandler struct {
	Ctx context.Context
	Dbc *postgres.DatabaseConnection
}

func (s *ScheduleRelationHandler) GetScheduleRelations(w http.ResponseWriter, r *http.Request) {
	schedulestypes, err := s.Dbc.GetScheduleRelations()
	if err != nil {
		httpsuite.WriteJSONError(w, "failed to get schedule types", http.StatusInternalServerError)
	}

	httpsuite.SendResponse(s.Ctx, w, "", http.StatusOK, &schedulestypes)
}

func (s *ScheduleRelationHandler) GetScheduleRelationByPersonnelID(w http.ResponseWriter, r *http.Request) {

	var input struct {
		ID string `json:"id" binding:"required"`
	}

	input.ID = chi.URLParam(r, "id")

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpsuite.SendResponse(s.Ctx, w, "failed to bind input", http.StatusBadRequest, &err)
		return
	}

	validationErr := httpsuite.IsRequestValid(input)
	if validationErr != nil {
		httpsuite.SendResponse(s.Ctx, w, "validation error", http.StatusBadRequest, validationErr)
		return
	}
	// personnel, err := pg.GetPersonnelSchedule(c.Query("name"))
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	return
	// }
	// c.JSON(http.StatusOK, personnel)
}

func (s *ScheduleRelationHandler) NewScheduleRelation(w http.ResponseWriter, r *http.Request) {

	// var input struct {
	// 	PersonnelName string `json:"personnel_name" binding:"required"`
	// 	ScheduleId    string `json:"schedule_id" binding:"required"`
	// }
	//
	// if err := c.ShouldBindJSON(&input); err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 	return
	// }

	// _, err := pg.Pool.Exec(pg.Ctx, "INSERT INTO personnel (name) VALUES ($1)", input.)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	return
	// }

	// c.JSON(http.StatusCreated, gin.H{"message": "personnel created"})
}
