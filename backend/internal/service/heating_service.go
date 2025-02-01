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
	DeleteAll() error
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

	// Base calculation
	baseHeatingTime := day.ShowerDuration*0.8 + 15
	temperatureFactor := (30 - day.AverageTemperature) / 3
	heatingTime := baseHeatingTime + temperatureFactor

	log.Printf("Base calculation: baseHeatingTime=%v, temperatureFactor=%v, initial heatingTime=%v",
		baseHeatingTime, temperatureFactor, heatingTime)

	recentHistory := s.storage.GetRecentEntries(5)
	log.Printf("Recent history entries: %d", len(recentHistory))

	if len(recentHistory) > 0 {
		lastEntry := recentHistory[0]
		var adjustment float64

		// If we have at least 2 entries, check for temperature swings
		if len(recentHistory) >= 2 {
			prevEntry := recentHistory[1]

			// Check if we switched from cold to hot or vice versa
			switchedFromColdToHot := prevEntry.Satisfaction < 5 && lastEntry.Satisfaction > 5
			switchedFromHotToCold := prevEntry.Satisfaction > 5 && lastEntry.Satisfaction < 5

			if switchedFromColdToHot || switchedFromHotToCold {
				// We overshot, take the midpoint between the two times
				midpoint := (prevEntry.HeatingTime + lastEntry.HeatingTime) / 2
				adjustment = midpoint - heatingTime
				log.Printf("Temperature swing detected, adjusting to midpoint: %v", midpoint)
			} else {
				// Count consecutive similar ratings
				consecutiveSimilar := 1
				lastSatisfaction := lastEntry.Satisfaction
				for i := 1; i < len(recentHistory); i++ {
					if recentHistory[i].Satisfaction != lastSatisfaction {
						break
					}
					consecutiveSimilar++
				}

				log.Printf("Found %d consecutive similar ratings of %v", consecutiveSimilar, lastSatisfaction)

				if lastSatisfaction < 5.0 {
					// Too cold - need more heating time
					coldness := float64(5.0 - lastSatisfaction)
					// Base adjustment - smaller steps to avoid overshooting
					adjustment = coldness * 2.0
					if consecutiveSimilar > 1 {
						// Increase more aggressively for persistent cold
						adjustment *= float64(consecutiveSimilar) * 0.5
					}
				} else if lastSatisfaction > 5.0 {
					// Too hot - need less heating time
					hotness := float64(lastSatisfaction - 5.0)
					// Base adjustment - respond quickly to too hot
					adjustment = -hotness * 3.0
					if consecutiveSimilar > 1 {
						// Decrease more aggressively for persistent heat
						adjustment *= float64(consecutiveSimilar) * 0.7
					}
				}
			}
		} else {
			// Only one entry, use simple adjustment
			if lastEntry.Satisfaction < 5.0 {
				adjustment = float64(5.0-lastEntry.Satisfaction) * 2.0
			} else if lastEntry.Satisfaction > 5.0 {
				adjustment = float64(5.0-lastEntry.Satisfaction) * 3.0
			}
		}

		// Apply the adjustment
		if adjustment != 0 {
			heatingTime += adjustment
			log.Printf("Applied adjustment of %v minutes based on feedback", adjustment)
		}
	}

	finalTime := clamp(heatingTime, 20, 80)
	log.Printf("Final heating time (after clamping): %v", finalTime)
	return finalTime
}

func (s *HeatingService) DeleteEntry(id string) error {
	return s.storage.DeleteEntry(id)
}

func (s *HeatingService) DeleteAll() error {
	return s.storage.DeleteAll()
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
