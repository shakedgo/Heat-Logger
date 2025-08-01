package main

import (
	router "heat-logger/internal/routes"
	"heat-logger/pkg/database"
	"log"
)

func main() {
	// Initialize database
	if err := database.InitDatabase(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Setup and start router
	r := router.SetupRouter()

	log.Println("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
