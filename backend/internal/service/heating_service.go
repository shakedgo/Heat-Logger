package service

import (
	"heat-logger/internal/models"
	"heat-logger/pkg/storage"
	"log"
)

type HeatingService struct {
	storage *storage.JSONStorage
}

func NewHeatingService(storage *storage.JSONStorage) *HeatingService {
	return &HeatingService{
		storage: storage,
	}
}

func (s *HeatingService) CalculateHeatingTime(day models.Day) float64 {
	log.Printf("Calculating heating time for input: %+v", day)

	baseHeatingTime := day.ShowerDuration + 10
	temperatureFactor := (30 - day.AverageTemperature) / 2
	heatingTime := baseHeatingTime + temperatureFactor

	log.Printf("Base calculation: baseHeatingTime=%v, temperatureFactor=%v, initial heatingTime=%v",
		baseHeatingTime, temperatureFactor, heatingTime)

	recentHistory := s.storage.GetRecentEntries(5)
	log.Printf("Recent history entries: %d", len(recentHistory))

	if len(recentHistory) > 0 {
		var totalAdjustment float64
		var weightSum float64

		for i, entry := range recentHistory {
			recencyWeight := float64(i+1) / float64(len(recentHistory))
			tempDiff := abs(entry.AverageTemperature - day.AverageTemperature)
			tempWeight := 1.0
			if tempDiff < 5 {
				tempWeight = 1.5
			}
			weight := recencyWeight * tempWeight

			var adjustment float64
			if entry.Satisfaction < 5 {
				adjustment = float64(5-entry.Satisfaction) * 2
			} else {
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

func (s *HeatingService) SaveFeedback(day models.Day) error {
	return s.storage.AddEntry(day)
}

func (s *HeatingService) GetHistory() []models.Day {
	return s.storage.GetHistory()
}

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
