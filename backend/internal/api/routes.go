package api

import (
	"fmt"
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
		fmt.Println("\nRegistering routes:")

		fmt.Println("GET /api/history")
		api.GET("/history", heatingHandler.GetHistory)

		fmt.Println("GET /api/history/export")
		api.GET("/history/export", heatingHandler.ExportHistory)

		fmt.Println("POST /api/history/delete")
		api.POST("/history/delete", heatingHandler.DeleteEntry)

		fmt.Println("POST /api/history/deleteall")
		api.POST("/history/deleteall", heatingHandler.DeleteAll)

		fmt.Println("POST /api/calculate")
		api.POST("/calculate", heatingHandler.Calculate)

		fmt.Println("POST /api/feedback")
		api.POST("/feedback", heatingHandler.SaveFeedback)

		fmt.Println("Routes registered successfully\n")
	}

	return r
}
