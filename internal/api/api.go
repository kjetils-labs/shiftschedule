package api

import (
	"context"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/shiftschedule/internal/api/routes"
)

func StartListening(ctx context.Context, sigs chan os.Signal) {
	go func(ctx context.Context, sigs chan os.Signal) {
		InitHttpServer(ctx)
		<-sigs
		ctx.Done()
	}(ctx, sigs)
}

func InitHttpServer(ctx context.Context) {
	router := gin.Default()
	setupRoutes(router)
	router.Run()
}

func setupRoutes(router *gin.Engine) {
	v1 := router.Group("/v1")
	v1.GET("/ping", routes.Ping())

	personnel := v1.Group("/personnel")
	personnel.GET("", routes.Ping(), routes.GetPersonnelAll())
	personnel.GET("/:personnelName", routes.GetPersonnelByName())
	personnel.POST("", routes.NewPersonnel())
	personnel.PUT("/:personelKey", routes.UpdatePersonnel())
	personnel.DELETE("/:personelKey", routes.DeletePersonnel())
}
