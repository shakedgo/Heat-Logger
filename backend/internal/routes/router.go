package router

import (
	"heat-logger/internal/config"
	"heat-logger/internal/handler"
	"heat-logger/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter(cfg *config.Config) *gin.Engine {
	r := gin.Default()

	// Configure CORS for frontend integration
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = cfg.CORS.AllowedOrigins
	corsConfig.AllowMethods = cfg.CORS.AllowedMethods
	corsConfig.AllowHeaders = cfg.CORS.AllowedHeaders
	corsConfig.AllowCredentials = true

	r.Use(cors.New(corsConfig))

	// Initialize services
	recordService := services.NewRecordService()
	useV2 := cfg.Prediction.Version != "v1"

	var predictor services.Predictor
	if useV2 {
		predictor = services.NewPredictionServiceV2(recordService, nil)
	} else {
		predictor = services.NewPredictionService(recordService) // v1 implements Predictor via shim
	}

	// Initialize handlers
	recordHandler := handler.NewRecordHandler(recordService, predictor)
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
