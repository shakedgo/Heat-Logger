package handlers

import (
	"fmt"
	"log"
	"net/http"

	"heat-logger/internal/models"
	"heat-logger/internal/service"

	"github.com/gin-gonic/gin"
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
	var day models.Day
	if err := c.ShouldBindJSON(&day); err != nil {
		log.Printf("Error binding JSON in Calculate: %v", err)
		log.Printf("Request body: %v", c.Request.Body)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid request format: %v", err),
		})
		return
	}

	log.Printf("Received calculation request for: %+v", day)
	heatingTime := h.service.CalculateHeatingTime(day)
	log.Printf("Calculated heating time: %v", heatingTime)

	c.JSON(http.StatusOK, gin.H{"heatingTime": heatingTime})
}

func (h *HeatingHandler) SaveFeedback(c *gin.Context) {
	var day models.Day
	if err := c.ShouldBindJSON(&day); err != nil {
		log.Printf("Error binding JSON in SaveFeedback: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid request format: %v", err),
		})
		return
	}

	if err := h.service.SaveFeedback(day); err != nil {
		log.Printf("Error saving feedback: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to save feedback: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
