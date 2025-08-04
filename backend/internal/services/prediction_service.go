package services

import (
	"math"
	"time"

	"heat-logger/internal/models"
)

// RecordServiceInterface defines the interface for record service operations needed by prediction service
type RecordServiceInterface interface {
	GetRecordsForPredictionByUser(userID string, limit int) ([]models.DailyRecord, error)
	GetGlobalRecordsForPrediction(excludeUserID string, limit int) ([]models.DailyRecord, error)
	GetRecordsForPrediction(limit int) ([]models.DailyRecord, error)
}

// PredictionService handles ML prediction logic
type PredictionService struct {
	recordService RecordServiceInterface
}

// NewPredictionService creates a new prediction service instance
func NewPredictionService(recordService *RecordService) *PredictionService {
	return &PredictionService{
		recordService: recordService,
	}
}

// PredictionRequest represents the input for heating time prediction
type PredictionRequest struct {
	UserID      string  `json:"userId" binding:"required"`
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

// PredictHeatingTime calculates the optimal heating time using hybrid user/global model
func (s *PredictionService) PredictHeatingTime(req *PredictionRequest) (*PredictionResponse, error) {
	// Get user-specific records
	userRecords, err := s.recordService.GetRecordsForPredictionByUser(req.UserID, 50)
	if err != nil {
		return nil, err
	}

	// Get global records (excluding this user to avoid duplication)
	globalRecords, err := s.recordService.GetGlobalRecordsForPrediction(req.UserID, 50)
	if err != nil {
		return nil, err
	}

	// Calculate hybrid prediction
	heatingTime := s.getCombinedPrediction(req, userRecords, globalRecords)

	return &PredictionResponse{
		HeatingTime: math.Round(heatingTime*10) / 10, // Round to 1 decimal place
	}, nil
}

// predictWithDefaults returns a prediction using default values when no historical data exists
func (s *PredictionService) predictWithDefaults(req *PredictionRequest) *PredictionResponse {
	// Base heating time calculation with improved default factors
	baseTime := 12.0      // Increased base heating time (was 8.0)
	durationFactor := 0.4 // More time per minute of shower (was 0.3)
	tempFactor := -0.15   // More temperature sensitivity (was -0.1)

	heatingTime := baseTime + (req.Duration * durationFactor) + (req.Temperature * tempFactor)

	// Ensure minimum heating time
	if heatingTime < 2.0 {
		heatingTime = 2.0
	}

	return &PredictionResponse{
		HeatingTime: math.Round(heatingTime*10) / 10,
	}
}

// getCombinedPrediction combines user-specific and global predictions using weighted average
func (s *PredictionService) getCombinedPrediction(req *PredictionRequest, userRecords, globalRecords []models.DailyRecord) float64 {
	// Calculate user weight based on amount of relevant data
	userWeight := s.calculateUserWeight(req, userRecords)
	globalWeight := 1.0 - userWeight

	// Calculate user-specific prediction
	var userPrediction float64
	if userWeight > 0 {
		userPrediction = s.calculatePredictionFromRecords(req, userRecords)
	}

	// Calculate global prediction
	globalPrediction := s.calculatePredictionFromRecords(req, globalRecords)

	// If no user data, return global prediction
	if userWeight == 0 {
		return globalPrediction
	}

	// If no global data, return user prediction or defaults
	if len(globalRecords) == 0 {
		if userWeight > 0 {
			return userPrediction
		}
		return s.predictWithDefaults(req).HeatingTime
	}

	// Combine predictions using weighted average
	finalPrediction := (userPrediction * userWeight) + (globalPrediction * globalWeight)

	// Ensure the prediction is within reasonable bounds
	if finalPrediction < 2.0 {
		return 2.0
	}
	if finalPrediction > 30.0 {
		return 30.0
	}

	return finalPrediction
}

// calculateUserWeight determines how much weight to give to user-specific data
func (s *PredictionService) calculateUserWeight(req *PredictionRequest, userRecords []models.DailyRecord) float64 {
	// Count relevant user records (similar conditions)
	relevantCount := 0
	for _, record := range userRecords {
		tempDiff := math.Abs(record.AverageTemperature - req.Temperature)
		durationDiff := math.Abs(record.ShowerDuration - req.Duration)

		// Count records with similar conditions
		if tempDiff <= 2.0 && durationDiff <= 3.0 {
			relevantCount++
		}
	}

	// Weight increases with relevant records, max weight at 10 records
	userWeight := math.Min(1.0, float64(relevantCount)/10.0)
	return userWeight
}

// calculatePredictionFromRecords calculates prediction from a set of records
func (s *PredictionService) calculatePredictionFromRecords(req *PredictionRequest, records []models.DailyRecord) float64 {
	// If no records, use defaults
	if len(records) == 0 {
		return s.predictWithDefaults(req).HeatingTime
	}

	// Use the existing prediction logic
	return s.calculatePrediction(req, records)
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

		// Calculate the adjustment needed based on user satisfaction using 2x² scaling
		var adjustment float64
		if record.Satisfaction != 50.0 {
			// Calculate distance from perfect satisfaction
			x := record.Satisfaction - 50.0
			normalizedX := x / 50.0

			// Apply 2x² scaling: f(x) = 2x² for more aggressive learning
			quadraticFactor := 2.0 * math.Pow(math.Abs(normalizedX), 2.0)

			// Increased base adjustment percentage for more aggressive learning
			baseAdjustmentPercent := 1.2 // Increased from 0.8 (50% more aggressive)

			// Apply pattern recognition boost
			coldBoost, hotBoost := s.detectExtremeFeedbackPattern(records)

			// Apply contextual learning boost based on progression
			contextualBoost := s.calculateContextualBoost(records, record.Satisfaction, x < 0)

			if x < 0 {
				// Cold feedback - increase heating time
				adjustment = quadraticFactor * (record.HeatingTime * baseAdjustmentPercent) * coldBoost * contextualBoost
			} else {
				// Hot feedback - decrease heating time
				adjustment = -quadraticFactor * (record.HeatingTime * baseAdjustmentPercent) * hotBoost * contextualBoost
			}
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

// detectExtremeFeedbackPattern detects consecutive extreme feedback patterns and returns boost factors
func (s *PredictionService) detectExtremeFeedbackPattern(records []models.DailyRecord) (float64, float64) {
	var consecutiveCold, consecutiveHot int
	var coldBoost, hotBoost float64 = 1.0, 1.0

	// Count consecutive extreme feedback (most recent first)
	for i := len(records) - 1; i >= 0; i-- {
		record := records[i]

		if record.Satisfaction < 30.0 {
			consecutiveCold++
			consecutiveHot = 0
		} else if record.Satisfaction > 70.0 {
			consecutiveHot++
			consecutiveCold = 0
		} else {
			// Reset counters for moderate feedback
			consecutiveCold = 0
			consecutiveHot = 0
		}

		// Apply pattern-based boost
		if consecutiveCold >= 3 {
			coldBoost = 1.5 + (float64(consecutiveCold) * 0.2) // 1.5x to 2.5x boost
		}
		if consecutiveHot >= 3 {
			hotBoost = 1.5 + (float64(consecutiveHot) * 0.2) // 1.5x to 2.5x boost
		}
	}

	return coldBoost, hotBoost
}

// calculateContextualBoost analyzes the learning progression to determine if we need more aggressive adjustments
func (s *PredictionService) calculateContextualBoost(records []models.DailyRecord, currentSatisfaction float64, isCold bool) float64 {
	if len(records) < 2 {
		return 1.0 // Not enough history for contextual analysis
	}

	// Get the last 3 records for progression analysis
	recentRecords := s.getRecentRecords(records, 3)
	if len(recentRecords) < 2 {
		return 1.0
	}

	// Analyze satisfaction progression
	var contextualBoost float64 = 1.0

	if isCold {
		// For cold feedback, check if satisfaction is improving or still low
		contextualBoost = s.analyzeColdProgression(recentRecords, currentSatisfaction)
	} else {
		// For hot feedback, check if satisfaction is improving or still high
		contextualBoost = s.analyzeHotProgression(recentRecords, currentSatisfaction)
	}

	return contextualBoost
}

// analyzeColdProgression determines if we need more aggressive heating adjustments
func (s *PredictionService) analyzeColdProgression(recentRecords []models.DailyRecord, currentSatisfaction float64) float64 {
	// If current satisfaction is still very low (< 40), we need to be more aggressive
	if currentSatisfaction < 40.0 {
		// Check if this is a pattern of low satisfaction
		lowSatisfactionCount := 0
		for _, record := range recentRecords {
			if record.Satisfaction < 40.0 {
				lowSatisfactionCount++
			}
		}

		// If we have multiple low satisfaction records, be more aggressive
		if lowSatisfactionCount >= 2 {
			// Calculate how much we've been adjusting
			adjustmentAggressiveness := s.calculateAdjustmentAggressiveness(recentRecords)

			// If we've been conservative, be more aggressive
			if adjustmentAggressiveness < 0.3 {
				return 3.0 // Triple the adjustment (increased from 2.0)
			} else if adjustmentAggressiveness < 0.5 {
				return 2.0 // Double the adjustment (increased from 1.5)
			}
		}

		// If satisfaction is extremely low (< 30), always be more aggressive
		if currentSatisfaction < 30.0 {
			return 2.5 // Increased from 1.8
		}
	}

	return 1.0
}

// analyzeHotProgression determines if we need more aggressive cooling adjustments
func (s *PredictionService) analyzeHotProgression(recentRecords []models.DailyRecord, currentSatisfaction float64) float64 {
	// If current satisfaction is still very high (> 60), we need to be more aggressive
	if currentSatisfaction > 60.0 {
		// Check if this is a pattern of high satisfaction
		highSatisfactionCount := 0
		for _, record := range recentRecords {
			if record.Satisfaction > 60.0 {
				highSatisfactionCount++
			}
		}

		// If we have multiple high satisfaction records, be more aggressive
		if highSatisfactionCount >= 2 {
			// Calculate how much we've been adjusting
			adjustmentAggressiveness := s.calculateAdjustmentAggressiveness(recentRecords)

			// If we've been conservative, be more aggressive
			if adjustmentAggressiveness < 0.3 {
				return 3.0 // Triple the adjustment (increased from 2.0)
			} else if adjustmentAggressiveness < 0.5 {
				return 2.0 // Double the adjustment (increased from 1.5)
			}
		}

		// If satisfaction is extremely high (> 70), always be more aggressive
		if currentSatisfaction > 70.0 {
			return 2.5 // Increased from 1.8
		}
	}

	return 1.0
}

// calculateAdjustmentAggressiveness measures how aggressively we've been adjusting
func (s *PredictionService) calculateAdjustmentAggressiveness(records []models.DailyRecord) float64 {
	if len(records) < 2 {
		return 0.0
	}

	var totalAdjustmentPercent float64
	var adjustmentCount int

	for i := 1; i < len(records); i++ {
		current := records[i]
		previous := records[i-1]

		// Calculate percentage change in heating time
		if previous.HeatingTime > 0 {
			adjustmentPercent := math.Abs(current.HeatingTime-previous.HeatingTime) / previous.HeatingTime
			totalAdjustmentPercent += adjustmentPercent
			adjustmentCount++
		}
	}

	if adjustmentCount > 0 {
		return totalAdjustmentPercent / float64(adjustmentCount)
	}

	return 0.0
}

// getRecentRecords returns the most recent N records
func (s *PredictionService) getRecentRecords(records []models.DailyRecord, count int) []models.DailyRecord {
	if len(records) <= count {
		return records
	}

	// Return the most recent records (they're already sorted by date)
	return records[len(records)-count:]
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
