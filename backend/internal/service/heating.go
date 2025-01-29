package service

import (
	"heat-logger/internal/models"
	"log"
)

type Storage interface {
	GetHistory() []models.Day
	AddEntry(day models.Day) error
	GetRecentEntries(limit int) []models.Day
	DeleteEntry(id string) error
}

type HeatingService struct {
	storage Storage
}

func NewHeatingService(storage Storage) *HeatingService {
	return &HeatingService{
		storage: storage,
	}
}

func (s *HeatingService) GetHistory() []models.Day {
	return s.storage.GetHistory()
}

func (s *HeatingService) SaveFeedback(day models.Day) error {
	return s.storage.AddEntry(day)
}

func (s *HeatingService) Calculate(duration float64, temperature float64) float64 {
	day := models.Day{
		ShowerDuration:     duration,
		AverageTemperature: temperature,
	}
	return s.CalculateHeatingTime(day)
}

func (s *HeatingService) CalculateHeatingTime(day models.Day) float64 {
	log.Printf("Calculating heating time for input: %+v", day)

	// Reduced base time calculation
	baseHeatingTime := day.ShowerDuration*0.8 + 15         // 80% of shower duration plus 15 minutes
	temperatureFactor := (30 - day.AverageTemperature) / 3 // Reduced temperature impact
	heatingTime := baseHeatingTime + temperatureFactor

	log.Printf("Base calculation: baseHeatingTime=%v, temperatureFactor=%v, initial heatingTime=%v",
		baseHeatingTime, temperatureFactor, heatingTime)

	recentHistory := s.storage.GetRecentEntries(5)
	log.Printf("Recent history entries: %d", len(recentHistory))

	if len(recentHistory) > 0 {
		var totalAdjustment float64
		var weightSum float64

		for i, entry := range recentHistory {
			// Newer entries get higher weights
			recencyWeight := float64(len(recentHistory)-i) / float64(len(recentHistory))

			// More weight to very similar temperatures
			tempDiff := abs(entry.AverageTemperature - day.AverageTemperature)
			tempWeight := 1.0
			if tempDiff < 2 {
				tempWeight = 2.0
			}
			weight := recencyWeight * tempWeight

			// Adjust based on satisfaction, with smaller adjustments
			var adjustment float64
			if entry.Satisfaction < 5 {
				adjustment = float64(5-entry.Satisfaction) * 2
			} else if entry.Satisfaction > 5 {
				adjustment = float64(entry.Satisfaction-5) * -2
			}

			totalAdjustment += adjustment * weight
			weightSum += weight

			log.Printf("Entry adjustment: recencyWeight=%v, tempWeight=%v, adjustment=%v",
				recencyWeight, tempWeight, adjustment)
		}

		if weightSum > 0 {
			heatingTime += totalAdjustment / weightSum
			log.Printf("Applied history adjustment: totalAdjustment=%v, weightSum=%v, final heatingTime=%v",
				totalAdjustment, weightSum, heatingTime)
		}
	}

	finalTime := clamp(heatingTime, 20, 80)
	log.Printf("Final heating time (after clamping): %v", finalTime)
	return finalTime
}

func (s *HeatingService) DeleteEntry(id string) error {
	return s.storage.DeleteEntry(id)
}

// Helper functions
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
