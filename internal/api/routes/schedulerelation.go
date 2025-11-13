package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shiftschedule/internal/clients/postgres"
)

func GetScheduleRelationBySchedule(pg *postgres.Postgres) gin.HandlerFunc {
	return func(c *gin.Context) {
		scheduleRelation, err := pg.GetSchedulePersonnel(c.Query("name"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, scheduleRelation)

	}
}

func GetScheduleRelationByPersonnel(pg *postgres.Postgres) gin.HandlerFunc {
	return func(c *gin.Context) {
		personnel, err := pg.GetPersonnelSchedule(c.Query("name"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, personnel)
	}
}

func NewScheduleRelation(pg *postgres.Postgres) gin.HandlerFunc {
	return func(c *gin.Context) {

		var input struct {
			PersonnelName string `json:"personnel_name" binding:"required"`
			ScheduleId    string `json:"schedule_id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// _, err := pg.Pool.Exec(pg.Ctx, "INSERT INTO personnel (name) VALUES ($1)", input.)
		// if err != nil {
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		// 	return
		// }

		c.JSON(http.StatusCreated, gin.H{"message": "personnel created"})
	}
}
