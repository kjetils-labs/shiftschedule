package api

import (
	"context"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/shiftschedule/internal/api/routes"
	"github.com/shiftschedule/internal/clients/postgres"
)

func StartListening(ctx context.Context, sigs chan os.Signal, pg *postgres.Postgres) {
	go func(ctx context.Context, sigs chan os.Signal) {
		InitHttpServer(ctx, pg)
		<-sigs
		ctx.Done()
	}(ctx, sigs)
}

func InitHttpServer(ctx context.Context, pg *postgres.Postgres) {
	router := gin.Default()
	setupRoutes(router, pg)
	router.Run()
}

func setupRoutes(router *gin.Engine, pg *postgres.Postgres) {
	v1 := router.Group("/v1")
	v1.GET("/ping", routes.Ping())

	personnel := v1.Group("/personnel")
	personnel.GET("", routes.Ping(), routes.GetPersonnelAll(pg))
	personnel.GET("/:personnelName", routes.GetPersonnelByName(pg))
	personnel.POST("", routes.NewPersonnel(pg))
	personnel.PUT("/:personelKey", routes.UpdatePersonnel(pg))
	personnel.DELETE("/:personelKey", routes.DeletePersonnel(pg))
}
