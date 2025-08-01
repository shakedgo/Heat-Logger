package router

import (
	"heat-logger/internal/handler"
	"heat-logger/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Configure CORS for frontend integration
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173", "http://localhost:3000", "http://127.0.0.1:5173"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	config.AllowCredentials = true

	r.Use(cors.New(config))

	// Initialize services
	recordService := services.NewRecordService()
	predictionService := services.NewPredictionService(recordService)

	// Initialize handlers
	recordHandler := handler.NewRecordHandler(recordService, predictionService)

	// API routes
	api := r.Group("/api")
	{
		// Heating time calculation
		api.POST("/calculate", recordHandler.CalculateHeatingTime)

		// Feedback submission
		api.POST("/feedback", recordHandler.SubmitFeedback)

		// History management
		api.GET("/history", recordHandler.GetHistory)
		api.POST("/history/delete", recordHandler.DeleteRecord)
		api.POST("/history/deleteall", recordHandler.DeleteAllRecords)
		api.GET("/history/export", recordHandler.ExportHistory)

		// Health check
		api.GET("/health", func(c *gin.Context) {
			c.String(200, "OK")
		})
	}

	return r
}
