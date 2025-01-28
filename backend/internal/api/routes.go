package api

import (
	"heat-logger/internal/api/handlers"
	"heat-logger/internal/api/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(heatingHandler *handlers.HeatingHandler) *gin.Engine {
	r := gin.Default()

	// Configure CORS
	r.Use(middleware.ConfigureCORS())

	// API routes
	api := r.Group("/api")
	{
		api.GET("/history", heatingHandler.GetHistory)
		api.POST("/calculate", heatingHandler.Calculate)
		api.POST("/feedback", heatingHandler.SaveFeedback)
	}

	return r
}
