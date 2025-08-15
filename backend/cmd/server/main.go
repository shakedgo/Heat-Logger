package main

import (
	"heat-logger/internal/config"
	router "heat-logger/internal/routes"
	"heat-logger/pkg/database"
	"log"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize database
	if err := database.InitDatabase(cfg); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Setup and start router
	r := router.SetupRouter(cfg)

	log.Printf("Using predictor version: %s", cfg.Prediction.Version)
	log.Printf("Starting server on %s", cfg.GetServerAddress())
	if err := r.Run(cfg.GetServerAddress()); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
