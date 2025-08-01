package router

import (
	"heat-logger/internal/handler"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	api := r.Group("/api")
	{
		api.GET("/users", handler.GetUsers)
		api.POST("/users", handler.CreateUser)
		api.GET("/health", handler.HealthCheck)
	}

	return r
}
