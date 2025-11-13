package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shiftschedule/internal/clients/postgres"
)

func GetPersonnelAll(pg *postgres.Postgres) gin.HandlerFunc {
	return func(c *gin.Context) {
		personnel, err := pg.GetPersonnel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, personnel)

	}
}

func GetPersonnelByName(pg *postgres.Postgres) gin.HandlerFunc {
	return func(c *gin.Context) {
		personnel, err := pg.GetPersonnelByName(c.Query("name"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, personnel)
	}
}

func NewPersonnel(pg *postgres.Postgres) gin.HandlerFunc {
	return func(c *gin.Context) {

		var input struct {
			Name string `json:"name" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err := pg.Pool.Exec(pg.Ctx, "INSERT INTO personnel (name) VALUES ($1)", input.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "personnel created"})
	}
}

func UpdatePersonnel(pg *postgres.Postgres) gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func DeletePersonnel(pg *postgres.Postgres) gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}
