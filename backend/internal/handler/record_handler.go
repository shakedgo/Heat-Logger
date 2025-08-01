package handler

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"time"

	"heat-logger/internal/models"
	"heat-logger/internal/services"

	"github.com/gin-gonic/gin"
)

// RecordHandler handles HTTP requests for daily records
type RecordHandler struct {
	recordService     *services.RecordService
	predictionService *services.PredictionService
}

// NewRecordHandler creates a new record handler instance
func NewRecordHandler(recordService *services.RecordService, predictionService *services.PredictionService) *RecordHandler {
	return &RecordHandler{
		recordService:     recordService,
		predictionService: predictionService,
	}
}

// CalculateHeatingTime handles POST /api/calculate
func (h *RecordHandler) CalculateHeatingTime(c *gin.Context) {
	var req services.PredictionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data: " + err.Error(),
		})
		return
	}

	// Validate input ranges
	if req.Duration < 1 || req.Duration > 60 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Shower duration must be between 1 and 60 minutes",
		})
		return
	}

	if req.Temperature < -50 || req.Temperature > 50 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Temperature must be between -50 and 50 degrees Celsius",
		})
		return
	}

	// Get prediction
	prediction, err := h.predictionService.PredictHeatingTime(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to calculate heating time: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, prediction)
}

// SubmitFeedback handles POST /api/feedback
func (h *RecordHandler) SubmitFeedback(c *gin.Context) {
	var record models.DailyRecord

	if err := c.ShouldBindJSON(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data: " + err.Error(),
		})
		return
	}

	// Validate required fields
	if record.ShowerDuration <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Shower duration must be greater than 0",
		})
		return
	}

	if record.HeatingTime <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Heating time must be greater than 0",
		})
		return
	}

	if record.Satisfaction < 1 || record.Satisfaction > 10 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Satisfaction rating must be between 1 and 10",
		})
		return
	}

	// Set date if not provided
	if record.Date.IsZero() {
		record.Date = time.Now()
	}

	// Create record
	err := h.recordService.CreateRecord(&record)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save feedback: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Feedback saved successfully",
	})
}

// GetHistory handles GET /api/history
func (h *RecordHandler) GetHistory(c *gin.Context) {
	records, err := h.recordService.GetAllRecords()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve history: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"history": records,
	})
}

// DeleteRecord handles POST /api/history/delete
func (h *RecordHandler) DeleteRecord(c *gin.Context) {
	var req struct {
		ID string `json:"id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data: " + err.Error(),
		})
		return
	}

	err := h.recordService.DeleteRecord(req.ID)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Record not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete record: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Record deleted successfully",
	})
}

// DeleteAllRecords handles POST /api/history/deleteall
func (h *RecordHandler) DeleteAllRecords(c *gin.Context) {
	err := h.recordService.DeleteAllRecords()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete all records: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "All records deleted successfully",
	})
}

// ExportHistory handles GET /api/history/export
func (h *RecordHandler) ExportHistory(c *gin.Context) {
	records, err := h.recordService.GetAllRecords()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve history: " + err.Error(),
		})
		return
	}

	// Set response headers for CSV download
	filename := "heating_history_" + time.Now().Format("2006-01-02") + ".csv"
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename="+filename)

	// Create CSV writer
	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// Write header
	header := []string{"Date", "Shower Duration", "Average Temperature", "Heating Time", "Satisfaction"}
	if err := writer.Write(header); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to write CSV header",
		})
		return
	}

	// Write data rows
	for _, record := range records {
		row := []string{
			record.Date.Format("2006-01-02 15:04:05"),
			strconv.FormatFloat(record.ShowerDuration, 'f', 1, 64),
			strconv.FormatFloat(record.AverageTemperature, 'f', 1, 64),
			strconv.FormatFloat(record.HeatingTime, 'f', 1, 64),
			strconv.FormatFloat(record.Satisfaction, 'f', 1, 64),
		}
		if err := writer.Write(row); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to write CSV data",
			})
			return
		}
	}
}
