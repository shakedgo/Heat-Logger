package services

import (
	"math"
	"time"

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

// SimilarRecord represents a record with similarity score
type SimilarRecord struct {
	Record     models.DailyRecord
	Similarity float64
	Weight     float64
}

// PredictHeatingTime calculates the optimal heating time based on input parameters
func (s *PredictionService) PredictHeatingTime(req *PredictionRequest) (*PredictionResponse, error) {
	// Get recent records for learning
	records, err := s.recordService.GetRecordsForPrediction(50) // Increased limit for better learning
	if err != nil {
		return nil, err
	}

	// If no historical data, use default values
	if len(records) == 0 {
		return s.predictWithDefaults(req), nil
	}

	// Calculate prediction using similarity-based learning
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

// calculatePrediction uses a target-based approach to find the optimal heating time.
// It calculates a "target time" for each similar record and averages them.
func (s *PredictionService) calculatePrediction(req *PredictionRequest, records []models.DailyRecord) float64 {
	similarRecords := s.findSimilarRecords(req, records)
	if len(similarRecords) == 0 {
		return s.predictWithDefaults(req).HeatingTime
	}

	var totalWeightedTargetTime float64
	var totalWeight float64

	for _, similarRecord := range similarRecords {
		record := similarRecord.Record
		weight := similarRecord.Weight

		// For perfect scores, apply decay if they are contradicted by newer records.
		if record.Satisfaction == 50.0 {
			decay := s.calculatePerfectScoreDecay(record, similarRecords)
			weight *= decay
		}

		// Calculate the adjustment needed based on user satisfaction.
		var adjustment float64
		if record.Satisfaction < 50.0 {
			// User was cold, so we need to increase the time.
			coldnessFactor := (50.0 - record.Satisfaction) / 49.0 // 0-1 scale
			adjustment = coldnessFactor * 4.0                     // Max +4 mins
		} else if record.Satisfaction > 50.0 {
			// User was hot, so we need to decrease the time.
			hotnessFactor := (record.Satisfaction - 50.0) / 50.0 // 0-1 scale
			adjustment = -hotnessFactor * 4.0                    // Max -4 mins
		}

		// The new target is the actual time from the record, plus the adjustment.
		targetTime := record.HeatingTime + adjustment

		totalWeightedTargetTime += targetTime * weight
		totalWeight += weight
	}

	// The final prediction is the weighted average of all target times.
	if totalWeight > 0 {
		finalPrediction := totalWeightedTargetTime / totalWeight
		// Ensure the prediction is within reasonable bounds.
		if finalPrediction < 2.0 {
			return 2.0
		}
		if finalPrediction > 30.0 {
			return 30.0
		}
		return finalPrediction
	}

	return s.predictWithDefaults(req).HeatingTime
}

// findSimilarRecords finds records with similar temperature and duration
func (s *PredictionService) findSimilarRecords(req *PredictionRequest, records []models.DailyRecord) []SimilarRecord {
	var similarRecords []SimilarRecord
	now := time.Now()

	for _, record := range records {
		// Check if temperature is within ±2°C
		tempDiff := math.Abs(record.AverageTemperature - req.Temperature)
		if tempDiff > 2.0 {
			continue
		}

		// Check if duration is within ±3 minutes
		durationDiff := math.Abs(record.ShowerDuration - req.Duration)
		if durationDiff > 3.0 {
			continue
		}

		// Calculate similarity score (0-1, where 1 is perfect match)
		tempSimilarity := 1.0 - (tempDiff / 2.0)         // 0-1 scale
		durationSimilarity := 1.0 - (durationDiff / 3.0) // 0-1 scale
		overallSimilarity := (tempSimilarity + durationSimilarity) / 2.0

		// Calculate recency weight (recent records get 2x weight)
		daysSince := now.Sub(record.Date).Hours() / 24.0
		recencyWeight := 1.0
		if daysSince <= 7.0 { // Within last week
			recencyWeight = 2.0
		} else if daysSince <= 30.0 { // Within last month
			recencyWeight = 1.5
		}

		// Calculate frequency weight (more similar records = higher confidence)
		frequencyWeight := s.calculateFrequencyWeight(req, records, record)

		// Combined weight
		totalWeight := overallSimilarity * recencyWeight * frequencyWeight

		similarRecords = append(similarRecords, SimilarRecord{
			Record:     record,
			Similarity: overallSimilarity,
			Weight:     totalWeight,
		})
	}

	return similarRecords
}

// calculateFrequencyWeight gives higher weight when there are more similar records
func (s *PredictionService) calculateFrequencyWeight(req *PredictionRequest, allRecords []models.DailyRecord, currentRecord models.DailyRecord) float64 {
	similarCount := 0
	totalCount := len(allRecords)
	var totalSimilarity float64

	// Count how many records have similar conditions and calculate total similarity
	for _, record := range allRecords {
		tempDiff := math.Abs(record.AverageTemperature - req.Temperature)
		durationDiff := math.Abs(record.ShowerDuration - req.Duration)

		if tempDiff <= 2.0 && durationDiff <= 3.0 {
			similarCount++
			// Calculate similarity for this record
			tempSimilarity := 1.0 - (tempDiff / 2.0)
			durationSimilarity := 1.0 - (durationDiff / 3.0)
			overallSimilarity := (tempSimilarity + durationSimilarity) / 2.0
			totalSimilarity += overallSimilarity
		}
	}

	// If we have many similar records, give higher weight (more confidence)
	if similarCount > 0 {
		// Consider both count and average similarity
		countFactor := float64(similarCount) / float64(totalCount)
		avgSimilarity := totalSimilarity / float64(similarCount)
		return 1.0 + (countFactor * avgSimilarity)
	}

	return 1.0
}

// calculatePerfectScoreDecay reduces the weight of perfect scores if they've been contradicted by subsequent attempts
func (s *PredictionService) calculatePerfectScoreDecay(perfectRecord models.DailyRecord, allSimilarRecords []SimilarRecord) float64 {
	// Find all records that attempted the same heating time after this perfect score
	var subsequentAttempts []models.DailyRecord

	for _, similarRecord := range allSimilarRecords {
		record := similarRecord.Record
		// Check if this record is after the perfect score and uses similar heating time (±0.2 minutes)
		if record.Date.After(perfectRecord.Date) &&
			math.Abs(record.HeatingTime-perfectRecord.HeatingTime) <= 0.2 {
			subsequentAttempts = append(subsequentAttempts, record)
		}
	}

	// If no subsequent attempts, no decay needed
	if len(subsequentAttempts) == 0 {
		return 1.0
	}

	// Calculate average satisfaction of subsequent attempts
	var totalSatisfaction float64
	for _, attempt := range subsequentAttempts {
		totalSatisfaction += attempt.Satisfaction
	}
	avgSatisfaction := totalSatisfaction / float64(len(subsequentAttempts))

	// If subsequent attempts are consistently worse than perfect (50), apply decay
	if avgSatisfaction < 50.0 && len(subsequentAttempts) >= 2 {
		// Calculate decay based on how much worse and how many attempts
		satisfactionDrop := 50.0 - avgSatisfaction
		attemptCount := float64(len(subsequentAttempts))

		// Decay formula: more attempts with lower satisfaction = more decay
		// Base decay of 0.5, additional decay based on satisfaction drop and attempt count
		decayFactor := 0.5 - (satisfactionDrop / 100.0) - (attemptCount * 0.1)

		// Ensure decay factor is between 0.1 and 1.0
		if decayFactor < 0.1 {
			decayFactor = 0.1
		}
		if decayFactor > 1.0 {
			decayFactor = 1.0
		}

		return decayFactor
	}

	// No significant decay needed
	return 1.0
}
