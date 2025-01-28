package main

import (
	"log"
	"path/filepath"

	"heat-logger/internal/api"
	"heat-logger/internal/api/handlers"
	"heat-logger/internal/service"
	"heat-logger/pkg/storage"
)

func main() {
	// Initialize storage
	dataFile := filepath.Join("data", "heating.json")
	store := storage.NewJSONStorage(dataFile)
	if err := store.Load(); err != nil {
		log.Fatal("Failed to load data:", err)
	}

	// Initialize service
	heatingService := service.NewHeatingService(store)

	// Initialize handlers
	heatingHandler := handlers.NewHeatingHandler(heatingService)

	// Setup router
	router := api.SetupRouter(heatingHandler)

	// Start server
	log.Fatal(router.Run(":8080"))
}
