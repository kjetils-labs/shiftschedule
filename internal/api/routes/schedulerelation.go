package routes

import (
	"net/http"

	"github.com/shiftschedule/internal/clients/postgres"
)

type ScheduleRelationHandler struct {
	pg *postgres.Postgres
}

func (s *ScheduleRelationHandler) GetScheduleRelationBySchedule(w http.ResponseWriter, r *http.Request) error {
	// scheduleRelation, err := pg.GetSchedulePersonnel(c.Query("name"))
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	return
	// }
	// c.JSON(http.StatusOK, scheduleRelation)

	return nil
}

func (s *ScheduleRelationHandler) GetScheduleRelationByPersonnel(w http.ResponseWriter, r *http.Request) error {
	// personnel, err := pg.GetPersonnelSchedule(c.Query("name"))
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	return
	// }
	// c.JSON(http.StatusOK, personnel)
	return nil
}

func (s *ScheduleRelationHandler) NewScheduleRelation(w http.ResponseWriter, r *http.Request) error {

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
	return nil
}
