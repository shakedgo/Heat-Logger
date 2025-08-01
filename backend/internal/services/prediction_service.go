package services

import (
	"math"

	"heat-logger/internal/models"
)

// PredictionService handles ML prediction logic
type PredictionService struct {
	recordService *RecordService
}

// NewPredictionService creates a new prediction service instance
func NewPredictionService(recordService *RecordService) *PredictionService {
	return &PredictionService{
		recordService: recordService,
	}
}

// PredictionRequest represents the input for heating time prediction
type PredictionRequest struct {
	Duration    float64 `json:"duration" binding:"required,min=1,max=60"`
	Temperature float64 `json:"temperature" binding:"required,min=-50,max=50"`
}

// PredictionResponse represents the prediction output
type PredictionResponse struct {
	HeatingTime float64 `json:"heatingTime"`
}

// PredictHeatingTime calculates the optimal heating time based on input parameters
func (s *PredictionService) PredictHeatingTime(req *PredictionRequest) (*PredictionResponse, error) {
	// Get recent records for learning
	records, err := s.recordService.GetRecordsForPrediction(10)
	if err != nil {
		return nil, err
	}

	// If no historical data, use default values
	if len(records) == 0 {
		return s.predictWithDefaults(req), nil
	}

	// Calculate prediction using simple linear regression
	heatingTime := s.calculatePrediction(req, records)

	return &PredictionResponse{
		HeatingTime: math.Round(heatingTime*10) / 10, // Round to 1 decimal place
	}, nil
}

// predictWithDefaults returns a prediction using default values when no historical data exists
func (s *PredictionService) predictWithDefaults(req *PredictionRequest) *PredictionResponse {
	// Base heating time calculation with default factors
	baseTime := 8.0       // Base heating time in minutes
	durationFactor := 0.3 // Minutes per minute of shower
	tempFactor := -0.1    // Minutes per degree Celsius (negative because higher temp = less heating needed)

	heatingTime := baseTime + (req.Duration * durationFactor) + (req.Temperature * tempFactor)

	// Ensure minimum heating time
	if heatingTime < 2.0 {
		heatingTime = 2.0
	}

	return &PredictionResponse{
		HeatingTime: math.Round(heatingTime*10) / 10,
	}
}

// calculatePrediction uses historical data to improve prediction accuracy
func (s *PredictionService) calculatePrediction(req *PredictionRequest, records []models.DailyRecord) float64 {
	if len(records) < 3 {
		return s.predictWithDefaults(req).HeatingTime
	}

	// Calculate average satisfaction from recent records
	var totalSatisfaction float64
	var validRecords int

	for _, record := range records {
		if record.Satisfaction > 0 {
			totalSatisfaction += record.Satisfaction
			validRecords++
		}
	}

	avgSatisfaction := 5.0 // Default to neutral
	if validRecords > 0 {
		avgSatisfaction = totalSatisfaction / float64(validRecords)
	}

	// Adjust factors based on satisfaction ratings
	baseTime := 8.0
	durationFactor := 0.3
	tempFactor := -0.1

	// Adjust based on average satisfaction
	if avgSatisfaction < 4.0 {
		// Users were generally unsatisfied (too cold), increase heating time
		baseTime += 2.0
		durationFactor += 0.1
	} else if avgSatisfaction > 7.0 {
		// Users were generally satisfied, fine-tune
		baseTime -= 1.0
		durationFactor -= 0.05
	}

	heatingTime := baseTime + (req.Duration * durationFactor) + (req.Temperature * tempFactor)

	// Ensure reasonable bounds
	if heatingTime < 2.0 {
		heatingTime = 2.0
	} else if heatingTime > 30.0 {
		heatingTime = 30.0
	}

	return heatingTime
}
