package handlers

import (
	"encoding/csv"
	"fmt"
	"heat-logger/internal/models"
	"heat-logger/internal/service"
	"net/http"
	"strconv"
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

func (h *HeatingHandler) ExportHistory(c *gin.Context) {
	history := h.service.GetHistory()

	// Set headers for CSV download
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment;filename=heating_history.csv")

	// Create CSV writer
	writer := csv.NewWriter(c.Writer)

	// Write header
	headers := []string{"Date", "Temperature (°C)", "Duration (min)", "Heating Time (min)", "Satisfaction"}
	if err := writer.Write(headers); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write CSV headers"})
		return
	}

	// Write data rows
	for _, entry := range history {
		row := []string{
			entry.Date.Format("2006-01-02 15:04:05"),
			fmt.Sprintf("%.1f", entry.AverageTemperature),
			fmt.Sprintf("%.1f", entry.ShowerDuration),
			fmt.Sprintf("%.1f", entry.HeatingTime),
			strconv.Itoa(entry.Satisfaction),
		}
		if err := writer.Write(row); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write CSV data"})
			return
		}
	}

	// Flush the writer
	writer.Flush()

	if err := writer.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate CSV"})
		return
	}
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

func (h *HeatingHandler) DeleteAll(c *gin.Context) {
	fmt.Println("DeleteAll handler called")

	if err := h.service.DeleteAll(); err != nil {
		fmt.Printf("Error in DeleteAll: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fmt.Println("Successfully deleted all records")
	c.JSON(http.StatusOK, gin.H{"message": "All history deleted successfully"})
}
