package handlers

import (
	"fmt"
	"heat-logger/internal/models"
	"heat-logger/internal/service"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type HeatingHandler struct {
	service *service.HeatingService
}

func NewHeatingHandler(service *service.HeatingService) *HeatingHandler {
	return &HeatingHandler{
		service: service,
	}
}

func (h *HeatingHandler) GetHistory(c *gin.Context) {
	history := h.service.GetHistory()
	c.JSON(http.StatusOK, history)
}

func (h *HeatingHandler) Calculate(c *gin.Context) {
	var request struct {
		Duration    float64 `json:"duration" binding:"required"`
		Temperature float64 `json:"temperature" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	heatingTime := h.service.Calculate(request.Duration, request.Temperature)
	c.JSON(http.StatusOK, gin.H{"heatingTime": heatingTime})
}

func (h *HeatingHandler) SaveFeedback(c *gin.Context) {
	var feedback models.FeedbackRequest
	if err := c.ShouldBindJSON(&feedback); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	day := models.Day{
		ID:                 uuid.New().String(),
		Date:               time.Now(),
		ShowerDuration:     feedback.ShowerDuration,
		AverageTemperature: feedback.AverageTemperature,
		Satisfaction:       feedback.Satisfaction,
		HeatingTime:        feedback.HeatingTime,
	}

	if err := h.service.SaveFeedback(day); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Feedback saved successfully"})
}

func (h *HeatingHandler) DeleteEntry(c *gin.Context) {
	fmt.Printf("Received delete request at path: %s\n", c.Request.URL.Path)

	var request struct {
		ID string `json:"id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		fmt.Printf("Error binding JSON: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: ID is required"})
		return
	}

	fmt.Printf("Attempting to delete entry with ID: %s\n", request.ID)

	if err := h.service.DeleteEntry(request.ID); err != nil {
		fmt.Printf("Error deleting entry: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fmt.Println("Successfully deleted entry")
	c.JSON(http.StatusOK, gin.H{"message": "Entry deleted successfully"})
}
