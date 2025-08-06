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
	globalRecords, err := s.recordService.GetGlobalRecordsForPrediction(req.UserID, 200) // Fetch more for clustering
	if err != nil {
		return nil, err
	}

	// Calculate hybrid prediction
	heatingTime := s.getCombinedPrediction(req, userRecords, globalRecords)

	return &PredictionResponse{
		HeatingTime: math.Round(heatingTime), // Round to whole minutes
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
	if heatingTime < 5.0 {
		heatingTime = 5.0
	}

	return &PredictionResponse{
		HeatingTime: math.Round(heatingTime),
	}
}

// getClusteredGlobalRecords filters global records to find a user archetype matching the request.
func (s *PredictionService) getClusteredGlobalRecords(req *PredictionRequest, globalRecords []models.DailyRecord) []models.DailyRecord {
	// Define archetypes based on request parameters
	isLongShower := req.Duration > 15
	isHotWeather := req.Temperature > 20
	isColdWeather := req.Temperature < 10

	var clusteredRecords []models.DailyRecord
	for _, record := range globalRecords {
		// Simple clustering: match records with similar characteristics
		match := true
		if isLongShower && record.ShowerDuration <= 15 {
			match = false
		}
		if !isLongShower && record.ShowerDuration > 15 {
			match = false
		}
		if isHotWeather && record.AverageTemperature <= 20 {
			match = false
		}
		if isColdWeather && record.AverageTemperature >= 10 {
			match = false
		}

		if match {
			clusteredRecords = append(clusteredRecords, record)
		}
	}

	// If no specific cluster is found, return all global records to avoid having no data.
	if len(clusteredRecords) < 10 {
		return globalRecords
	}
	return clusteredRecords
}

// getCombinedPrediction combines user-specific and global predictions using weighted average
func (s *PredictionService) getCombinedPrediction(req *PredictionRequest, userRecords, globalRecords []models.DailyRecord) float64 {
	userWeight := s.calculateUserWeight(req, userRecords)
	globalWeight := 1.0 - userWeight

	var userPrediction float64
	if userWeight > 0 {
		userPrediction = s.calculatePredictionFromRecords(req, userRecords, len(userRecords))
	}

	// IMPROVEMENT 4: Use a clustered global model for more relevant predictions
	clusteredGlobalRecords := s.getClusteredGlobalRecords(req, globalRecords)
	globalPrediction := s.calculatePredictionFromRecords(req, clusteredGlobalRecords, len(clusteredGlobalRecords))

	if userWeight == 0 {
		return globalPrediction
	}

	if len(globalRecords) == 0 {
		if userWeight > 0 {
			return userPrediction
		}
		return s.predictWithDefaults(req).HeatingTime
	}

	finalPrediction := (userPrediction * userWeight) + (globalPrediction * globalWeight)

	// Ensure the prediction is within reasonable bounds
	if finalPrediction < 5.0 {
		return 5.0
	}
	if finalPrediction > 120.0 {
		return 120.0
	}

	return finalPrediction
}

// calculateUserWeight determines how much weight to give to user-specific data
func (s *PredictionService) calculateUserWeight(req *PredictionRequest, userRecords []models.DailyRecord) float64 {
	relevantCount := 0
	for _, record := range userRecords {
		tempDiff := math.Abs(record.AverageTemperature - req.Temperature)
		durationDiff := math.Abs(record.ShowerDuration - req.Duration)
		if tempDiff <= 2.0 && durationDiff <= 3.0 {
			relevantCount++
		}
	}
	return math.Min(1.0, float64(relevantCount)/10.0)
}

// calculatePredictionFromRecords calculates prediction from a set of records
func (s *PredictionService) calculatePredictionFromRecords(req *PredictionRequest, records []models.DailyRecord, totalRecordCount int) float64 {
	if len(records) == 0 {
		return s.predictWithDefaults(req).HeatingTime
	}
	return s.calculatePrediction(req, records, totalRecordCount)
}

// calculateDynamicLearningRate calculates a dynamic learning rate.
// The learning rate is higher for newer models (fewer records) and when feedback is far from perfect.
func (s *PredictionService) calculateDynamicLearningRate(satisfaction float64, recordCount int) float64 {
	// Start with a higher learning rate and decrease it as the model gains more data (confidence).
	confidenceFactor := 1.0 - math.Min(1.0, float64(recordCount)/30.0)*0.7 // From 1.0 down to 0.3
	// Increase learning rate based on how far the satisfaction is from the perfect score of 50.
	satisfactionFactor := 1.0 + math.Abs(satisfaction-50.0)/50.0 // Ranges from 1.0 to 2.0
	// Combine factors for the final dynamic rate.
	learningRate := confidenceFactor * satisfactionFactor
	// Clamp the rate to prevent extreme adjustments.
	return math.Max(0.2, math.Min(learningRate, 2.0))
}

// calculatePrediction uses a target-based approach to find the optimal heating time.
func (s *PredictionService) calculatePrediction(req *PredictionRequest, records []models.DailyRecord, totalRecordCount int) float64 {
	similarRecords := s.findSimilarRecords(req, records)
	if len(similarRecords) == 0 {
		return s.predictWithDefaults(req).HeatingTime
	}

	// IMPROVEMENT: Check if we're stuck in a pattern of poor predictions
	if s.isStuckInPattern(records) {
		return s.handleStuckPattern(records)
	}

	// IMPROVEMENT: Find weighted success anchors instead of just the last one
	successAnchors := s.findWeightedSuccessAnchors(records)

	var totalWeightedTargetTime float64
	var totalWeight float64

	for _, similarRecord := range similarRecords {
		record := similarRecord.Record
		weight := similarRecord.Weight

		if record.Satisfaction == 50.0 {
			decay := s.calculatePerfectScoreDecay(record, similarRecords)
			weight *= decay
		}

		var adjustment float64
		if record.Satisfaction != 50.0 {
			x := record.Satisfaction - 50.0
			normalizedX := x / 50.0

			quadraticFactor := 2.0 * math.Pow(math.Abs(normalizedX), 2.0)
			baseAdjustmentPercent := s.calculateDynamicLearningRate(record.Satisfaction, totalRecordCount)

			coldBoost, hotBoost := s.detectExtremeFeedbackPattern(records)
			contextualBoost := s.calculateContextualBoost(records, record.Satisfaction, x < 0)

			// IMPROVEMENT: Refined overshoot mechanism.
			baseOvershoot := 1.0 + (math.Abs(normalizedX) * 0.4)
			// IMPROVEMENT: Disable overshoot for any satisfaction > 50 to encourage fine-tuning.
			if record.Satisfaction > 50 {
				baseOvershoot = 1.0
			}
			dampeningFactor := 1.0 / (1.0 + (float64(len(similarRecords)) / 5.0))
			effectiveOvershoot := 1.0 + (baseOvershoot-1.0)*dampeningFactor

			if x < 0 {
				effectiveOvershoot *= 1.1
			}
			overshootFactor := math.Min(effectiveOvershoot, 1.4)

			if x < 0 {
				adjustment = quadraticFactor * (record.HeatingTime * baseAdjustmentPercent) * coldBoost * contextualBoost
			} else {
				adjustment = -quadraticFactor * (record.HeatingTime * baseAdjustmentPercent) * hotBoost * contextualBoost
				// IMPROVEMENT: More aggressive dampening for consecutive hot feedback
				consecutiveHotCount := s.countConsecutiveHotFeedback(records)
				if consecutiveHotCount >= 2 {
					// More aggressive reduction for consecutive hot feedback
					adjustment *= 0.4 + (0.2 * float64(consecutiveHotCount)) // 0.4 to 1.0 based on consecutive count
				} else {
					// Standard dampening for single hot feedback
					adjustment *= 0.6
				}
			}
			adjustment *= overshootFactor
		}

		targetTime := record.HeatingTime + adjustment
		totalWeightedTargetTime += targetTime * weight
		totalWeight += weight
	}

	if totalWeight > 0 {
		finalPrediction := totalWeightedTargetTime / totalWeight

		// IMPROVEMENT: Apply intelligent success anchor logic
		if len(successAnchors) > 0 {
			finalPrediction = s.applySuccessAnchorLogic(finalPrediction, successAnchors)
		}

		if finalPrediction < 5.0 {
			return 5.0
		}
		if finalPrediction > 120.0 {
			return 120.0
		}
		return finalPrediction
	}

	return s.predictWithDefaults(req).HeatingTime
}

// detectExtremeFeedbackPattern detects consecutive extreme feedback patterns and returns boost factors
func (s *PredictionService) detectExtremeFeedbackPattern(records []models.DailyRecord) (float64, float64) {
	var consecutiveCold, consecutiveHot int
	var coldBoost, hotBoost float64 = 1.0, 1.0

	for i := len(records) - 1; i >= 0; i-- {
		record := records[i]
		if record.Satisfaction < 30.0 {
			consecutiveCold++
			consecutiveHot = 0
		} else if record.Satisfaction > 70.0 {
			consecutiveHot++
			consecutiveCold = 0
		} else {
			consecutiveCold = 0
			consecutiveHot = 0
		}

		if consecutiveCold >= 3 {
			coldBoost = 1.5 + (float64(consecutiveCold) * 0.2)
		}
		if consecutiveHot >= 3 {
			hotBoost = 1.5 + (float64(consecutiveHot) * 0.2)
		}
	}
	return coldBoost, hotBoost
}

// calculateContextualBoost analyzes the learning progression to determine if we need more aggressive adjustments
func (s *PredictionService) calculateContextualBoost(records []models.DailyRecord, currentSatisfaction float64, isCold bool) float64 {
	if len(records) < 2 {
		return 1.0
	}
	recentRecords := s.getRecentRecords(records, 3)
	if len(recentRecords) < 2 {
		return 1.0
	}

	var contextualBoost float64 = 1.0
	if isCold {
		contextualBoost = s.analyzeColdProgression(recentRecords, currentSatisfaction)
	} else {
		contextualBoost = s.analyzeHotProgression(recentRecords, currentSatisfaction)
	}
	return contextualBoost
}

// analyzeColdProgression determines if we need more aggressive heating adjustments
func (s *PredictionService) analyzeColdProgression(recentRecords []models.DailyRecord, currentSatisfaction float64) float64 {
	if currentSatisfaction < 40.0 {
		lowSatisfactionCount := 0
		for _, record := range recentRecords {
			if record.Satisfaction < 40.0 {
				lowSatisfactionCount++
			}
		}
		if lowSatisfactionCount >= 2 {
			adjustmentAggressiveness := s.calculateAdjustmentAggressiveness(recentRecords)
			if adjustmentAggressiveness < 0.3 {
				return 3.0
			} else if adjustmentAggressiveness < 0.5 {
				return 2.0
			}
		}
		if currentSatisfaction < 30.0 {
			return 2.5
		}
	}
	return 1.0
}

// analyzeHotProgression determines if we need more aggressive cooling adjustments
func (s *PredictionService) analyzeHotProgression(recentRecords []models.DailyRecord, currentSatisfaction float64) float64 {
	if currentSatisfaction > 60.0 {
		highSatisfactionCount := 0
		for _, record := range recentRecords {
			if record.Satisfaction > 60.0 {
				highSatisfactionCount++
			}
		}
		if highSatisfactionCount >= 2 {
			adjustmentAggressiveness := s.calculateAdjustmentAggressiveness(recentRecords)
			if adjustmentAggressiveness < 0.3 {
				return 3.0
			} else if adjustmentAggressiveness < 0.5 {
				return 2.0
			}
		}
		if currentSatisfaction > 70.0 {
			return 2.5
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
	return records[len(records)-count:]
}

// WeightedSuccessAnchor represents a successful prediction with its weight
type WeightedSuccessAnchor struct {
	Record models.DailyRecord
	Weight float64
}

// IMPROVEMENT: Find multiple weighted success anchors instead of just the last one
func (s *PredictionService) findWeightedSuccessAnchors(records []models.DailyRecord) []WeightedSuccessAnchor {
	var anchors []WeightedSuccessAnchor
	now := time.Now()

	// Find all records with satisfaction > 55 (lowered threshold to include more hot feedback)
	for i := len(records) - 1; i >= 0; i-- {
		record := records[i]
		if record.Satisfaction > 55 {
			// Calculate weight based on recency and satisfaction level
			daysSince := now.Sub(record.Date).Hours() / 24.0
			recencyWeight := math.Exp(-0.1 * daysSince)             // Decay over ~10 days
			satisfactionWeight := (record.Satisfaction - 55) / 45.0 // 0-1 scale for 55-100
			totalWeight := recencyWeight * (1.0 + satisfactionWeight)

			anchors = append(anchors, WeightedSuccessAnchor{
				Record: record,
				Weight: totalWeight,
			})

			// Limit to top 3 most recent successes
			if len(anchors) >= 3 {
				break
			}
		}
	}

	return anchors
}

// IMPROVEMENT: Intelligent success anchor logic that uses successful predictions as starting points
func (s *PredictionService) applySuccessAnchorLogic(calculatedPrediction float64, successAnchors []WeightedSuccessAnchor) float64 {
	if len(successAnchors) == 0 {
		return calculatedPrediction
	}

	// Calculate weighted average of success anchors
	var totalWeightedTime float64
	var totalWeight float64

	for _, anchor := range successAnchors {
		// Apply graduated adjustment based on satisfaction level
		adjustedTime := s.applyGraduatedAdjustment(anchor.Record)
		totalWeightedTime += adjustedTime * anchor.Weight
		totalWeight += anchor.Weight
	}

	if totalWeight > 0 {
		anchorBasedPrediction := totalWeightedTime / totalWeight

		// Blend the anchor-based prediction with the calculated prediction
		// Give more weight to anchor when we have strong success history
		anchorInfluence := math.Min(0.7, totalWeight) // Max 70% influence
		calculatedInfluence := 1.0 - anchorInfluence

		return (anchorBasedPrediction * anchorInfluence) + (calculatedPrediction * calculatedInfluence)
	}

	return calculatedPrediction
}

// IMPROVEMENT: Apply graduated adjustments based on satisfaction level for any hot feedback
func (s *PredictionService) applyGraduatedAdjustment(record models.DailyRecord) float64 {
	satisfaction := record.Satisfaction
	heatingTime := record.HeatingTime

	// Apply different reduction percentages based on how "hot" the feedback was
	if satisfaction >= 85 {
		// Very hot - reduce by 25-30%
		return heatingTime * 0.75
	} else if satisfaction >= 80 {
		// Hot - reduce by 20-25%
		return heatingTime * 0.80
	} else if satisfaction >= 75 {
		// Moderately hot - reduce by 15-20%
		return heatingTime * 0.83
	} else if satisfaction >= 65 {
		// Slightly hot - reduce by 10-15%
		return heatingTime * 0.87
	} else if satisfaction >= 60 {
		// Warm - reduce by 7-10%
		return heatingTime * 0.92
	} else if satisfaction >= 55 {
		// Just above perfect - reduce by 3-5%
		return heatingTime * 0.97
	} else {
		// Should not reach here, but return original time
		return heatingTime
	}
}

// IMPROVEMENT: Detect when we're stuck in a pattern of similar poor predictions
func (s *PredictionService) isStuckInPattern(records []models.DailyRecord) bool {
	if len(records) < 4 {
		return false
	}

	// Get the last 4 records
	recentRecords := s.getRecentRecords(records, 4)

	// Check if all recent records have similar heating times and poor satisfaction
	var avgHeatingTime, avgSatisfaction float64
	for _, record := range recentRecords {
		avgHeatingTime += record.HeatingTime
		avgSatisfaction += record.Satisfaction
	}
	avgHeatingTime /= float64(len(recentRecords))
	avgSatisfaction /= float64(len(recentRecords))

	// Check if we're stuck: similar heating times, consistently poor satisfaction
	heatingTimeVariance := 0.0
	satisfactionBelowThreshold := 0

	for _, record := range recentRecords {
		heatingTimeVariance += math.Pow(record.HeatingTime-avgHeatingTime, 2)
		if record.Satisfaction < 50 {
			satisfactionBelowThreshold++
		}
	}

	heatingTimeVariance /= float64(len(recentRecords))

	// We're stuck if heating times are similar (low variance) and satisfaction is consistently poor
	return heatingTimeVariance < 4.0 && satisfactionBelowThreshold >= 3
}

// IMPROVEMENT: Handle stuck patterns by making a larger strategic adjustment
func (s *PredictionService) handleStuckPattern(records []models.DailyRecord) float64 {
	recentRecords := s.getRecentRecords(records, 4)

	// Calculate average of recent attempts
	var avgHeatingTime, avgSatisfaction float64
	for _, record := range recentRecords {
		avgHeatingTime += record.HeatingTime
		avgSatisfaction += record.Satisfaction
	}
	avgHeatingTime /= float64(len(recentRecords))
	avgSatisfaction /= float64(len(recentRecords))

	// Make a strategic jump based on how far we are from perfect
	if avgSatisfaction < 30 {
		// Very cold - increase by 50%
		return avgHeatingTime * 1.5
	} else if avgSatisfaction < 45 {
		// Cold - increase by 30%
		return avgHeatingTime * 1.3
	} else if avgSatisfaction > 70 {
		// Hot - decrease by 25%
		return avgHeatingTime * 0.75
	} else if avgSatisfaction > 55 {
		// Slightly hot - decrease by 15%
		return avgHeatingTime * 0.85
	}

	// Default: make a moderate adjustment
	return avgHeatingTime * 1.2
}

// IMPROVEMENT: Count consecutive hot feedback to make more aggressive adjustments
func (s *PredictionService) countConsecutiveHotFeedback(records []models.DailyRecord) int {
	consecutiveCount := 0

	// Count backwards from most recent record
	for i := len(records) - 1; i >= 0; i-- {
		if records[i].Satisfaction > 50 {
			consecutiveCount++
		} else {
			break // Stop at first non-hot feedback
		}
	}

	return consecutiveCount
}

// findSimilarRecords finds records with similar temperature and duration
func (s *PredictionService) findSimilarRecords(req *PredictionRequest, records []models.DailyRecord) []SimilarRecord {
	var similarRecords []SimilarRecord
	now := time.Now()

	for _, record := range records {
		tempDiff := math.Abs(record.AverageTemperature - req.Temperature)
		if tempDiff > 2.0 {
			continue
		}
		durationDiff := math.Abs(record.ShowerDuration - req.Duration)
		if durationDiff > 3.0 {
			continue
		}

		tempSimilarity := 1.0 - (tempDiff / 2.0)
		durationSimilarity := 1.0 - (durationDiff / 3.0)
		overallSimilarity := (tempSimilarity + durationSimilarity) / 2.0

		// Use continuous time-decay for recency weight.
		daysSince := now.Sub(record.Date).Hours() / 24.0
		decayConstant := 0.023 // Halves weight roughly every 30 days.
		recencyWeight := math.Exp(-decayConstant * daysSince)

		frequencyWeight := s.calculateFrequencyWeight(req, records, record)
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
	for _, record := range allRecords {
		tempDiff := math.Abs(record.AverageTemperature - req.Temperature)
		durationDiff := math.Abs(record.ShowerDuration - req.Duration)
		if tempDiff <= 2.0 && durationDiff <= 3.0 {
			similarCount++
			tempSimilarity := 1.0 - (tempDiff / 2.0)
			durationSimilarity := 1.0 - (durationDiff / 3.0)
			overallSimilarity := (tempSimilarity + durationSimilarity) / 2.0
			totalSimilarity += overallSimilarity
		}
	}
	if similarCount > 0 {
		countFactor := float64(similarCount) / float64(totalCount)
		avgSimilarity := totalSimilarity / float64(similarCount)
		return 1.0 + (countFactor * avgSimilarity)
	}
	return 1.0
}

// calculatePerfectScoreDecay reduces the weight of perfect scores if they've been contradicted by subsequent attempts
func (s *PredictionService) calculatePerfectScoreDecay(perfectRecord models.DailyRecord, allSimilarRecords []SimilarRecord) float64 {
	var subsequentAttempts []models.DailyRecord
	for _, similarRecord := range allSimilarRecords {
		record := similarRecord.Record
		if record.Date.After(perfectRecord.Date) &&
			math.Abs(record.HeatingTime-perfectRecord.HeatingTime) <= 0.2 {
			subsequentAttempts = append(subsequentAttempts, record)
		}
	}

	if len(subsequentAttempts) == 0 {
		return 1.0
	}

	var totalSatisfaction float64
	for _, attempt := range subsequentAttempts {
		totalSatisfaction += attempt.Satisfaction
	}
	avgSatisfaction := totalSatisfaction / float64(len(subsequentAttempts))

	if avgSatisfaction < 50.0 && len(subsequentAttempts) >= 2 {
		satisfactionDrop := 50.0 - avgSatisfaction
		attemptCount := float64(len(subsequentAttempts))
		decayFactor := 0.5 - (satisfactionDrop / 100.0) - (attemptCount * 0.1)

		if decayFactor < 0.1 {
			return 0.1
		}
		if decayFactor > 1.0 {
			return 1.0
		}
		return decayFactor
	}
	return 1.0
}
