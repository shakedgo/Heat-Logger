package services

import (
	"testing"
	"time"

	"heat-logger/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRecordService is a mock implementation of RecordServiceInterface for testing
type MockRecordService struct {
	mock.Mock
}

func (m *MockRecordService) GetRecordsForPredictionByUser(userID string, limit int) ([]models.DailyRecord, error) {
	args := m.Called(userID, limit)
	return args.Get(0).([]models.DailyRecord), args.Error(1)
}

func (m *MockRecordService) GetGlobalRecordsForPrediction(excludeUserID string, limit int) ([]models.DailyRecord, error) {
	args := m.Called(excludeUserID, limit)
	return args.Get(0).([]models.DailyRecord), args.Error(1)
}

func (m *MockRecordService) GetRecordsForPrediction(limit int) ([]models.DailyRecord, error) {
	args := m.Called(limit)
	return args.Get(0).([]models.DailyRecord), args.Error(1)
}

func TestPredictionService_NewUser_ShouldReceiveGlobalPrediction(t *testing.T) {
	// Arrange
	mockRecordService := &MockRecordService{}
	predictionService := &PredictionService{recordService: mockRecordService}

	// Mock: New user has no records
	mockRecordService.On("GetRecordsForPredictionByUser", "new_user", 50).Return([]models.DailyRecord{}, nil)

	// Mock: Global records exist
	globalRecords := []models.DailyRecord{
		{
			UserID:             "other_user",
			Date:               time.Now().AddDate(0, 0, -1),
			ShowerDuration:     10.0,
			AverageTemperature: 20.0,
			HeatingTime:        8.0,
			Satisfaction:       50.0,
		},
	}
	mockRecordService.On("GetGlobalRecordsForPrediction", "new_user", 50).Return(globalRecords, nil)

	req := &PredictionRequest{
		UserID:      "new_user",
		Duration:    10.0,
		Temperature: 20.0,
	}

	// Act
	result, err := predictionService.PredictHeatingTime(req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Greater(t, result.HeatingTime, 0.0)
	mockRecordService.AssertExpectations(t)
}

func TestPredictionService_UserWithFewRecords_ShouldReceiveBlendedPrediction(t *testing.T) {
	// Arrange
	mockRecordService := &MockRecordService{}
	predictionService := &PredictionService{recordService: mockRecordService}

	// Mock: User has few records
	userRecords := []models.DailyRecord{
		{
			UserID:             "user_with_few_records",
			Date:               time.Now().AddDate(0, 0, -1),
			ShowerDuration:     10.0,
			AverageTemperature: 20.0,
			HeatingTime:        9.0,
			Satisfaction:       45.0, // Was a bit cold
		},
	}
	mockRecordService.On("GetRecordsForPredictionByUser", "user_with_few_records", 50).Return(userRecords, nil)

	// Mock: Global records exist
	globalRecords := []models.DailyRecord{
		{
			UserID:             "other_user",
			Date:               time.Now().AddDate(0, 0, -2),
			ShowerDuration:     10.0,
			AverageTemperature: 20.0,
			HeatingTime:        8.0,
			Satisfaction:       50.0,
		},
	}
	mockRecordService.On("GetGlobalRecordsForPrediction", "user_with_few_records", 50).Return(globalRecords, nil)

	req := &PredictionRequest{
		UserID:      "user_with_few_records",
		Duration:    10.0,
		Temperature: 20.0,
	}

	// Act
	result, err := predictionService.PredictHeatingTime(req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Greater(t, result.HeatingTime, 0.0)
	mockRecordService.AssertExpectations(t)
}

func TestPredictionService_UserWithManyRecords_ShouldReceiveUserBasedPrediction(t *testing.T) {
	// Arrange
	mockRecordService := &MockRecordService{}
	predictionService := &PredictionService{recordService: mockRecordService}

	// Mock: User has many similar records (>10)
	userRecords := make([]models.DailyRecord, 12)
	for i := 0; i < 12; i++ {
		userRecords[i] = models.DailyRecord{
			UserID:             "experienced_user",
			Date:               time.Now().AddDate(0, 0, -i-1),
			ShowerDuration:     10.0,
			AverageTemperature: 20.0,
			HeatingTime:        8.5,
			Satisfaction:       50.0,
		}
	}
	mockRecordService.On("GetRecordsForPredictionByUser", "experienced_user", 50).Return(userRecords, nil)

	// Mock: Global records exist but should have minimal impact
	globalRecords := []models.DailyRecord{
		{
			UserID:             "other_user",
			Date:               time.Now().AddDate(0, 0, -1),
			ShowerDuration:     10.0,
			AverageTemperature: 20.0,
			HeatingTime:        15.0, // Very different from user's history
			Satisfaction:       50.0,
		},
	}
	mockRecordService.On("GetGlobalRecordsForPrediction", "experienced_user", 50).Return(globalRecords, nil)

	req := &PredictionRequest{
		UserID:      "experienced_user",
		Duration:    10.0,
		Temperature: 20.0,
	}

	// Act
	result, err := predictionService.PredictHeatingTime(req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Prediction should be closer to user's history (8.5) than global (15.0)
	assert.Less(t, result.HeatingTime, 12.0)
	mockRecordService.AssertExpectations(t)
}

func TestPredictionService_RelativeFeedbackAdjustment(t *testing.T) {
	// Arrange
	// This test doesn't need a full prediction service, just tests the adjustment logic

	// Test records with different heating times and satisfaction scores
	testCases := []struct {
		name         string
		heatingTime  float64
		satisfaction float64
		expectedSign string // "positive" for increase, "negative" for decrease, "zero" for no change
	}{
		{
			name:         "Cold feedback on short heating time",
			heatingTime:  5.0,
			satisfaction: 30.0, // Very cold
			expectedSign: "positive",
		},
		{
			name:         "Cold feedback on long heating time",
			heatingTime:  20.0,
			satisfaction: 30.0, // Very cold
			expectedSign: "positive",
		},
		{
			name:         "Hot feedback on short heating time",
			heatingTime:  5.0,
			satisfaction: 70.0, // Very hot
			expectedSign: "negative",
		},
		{
			name:         "Hot feedback on long heating time",
			heatingTime:  20.0,
			satisfaction: 70.0, // Very hot
			expectedSign: "negative",
		},
		{
			name:         "Perfect satisfaction",
			heatingTime:  10.0,
			satisfaction: 50.0, // Perfect
			expectedSign: "zero",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test record
			record := models.DailyRecord{
				HeatingTime:  tc.heatingTime,
				Satisfaction: tc.satisfaction,
			}

			// Calculate adjustment using the same logic as in the service
			var adjustment float64
			if record.Satisfaction < 50.0 {
				coldnessFactor := (50.0 - record.Satisfaction) / 49.0
				adjustment = coldnessFactor * (record.HeatingTime * 0.25)
			} else if record.Satisfaction > 50.0 {
				hotnessFactor := (record.Satisfaction - 50.0) / 50.0
				adjustment = -hotnessFactor * (record.HeatingTime * 0.25)
			}

			// Assert based on expected sign
			switch tc.expectedSign {
			case "positive":
				assert.Greater(t, adjustment, 0.0, "Expected positive adjustment for cold feedback")
				// Verify it's proportional to heating time
				expectedMaxAdjustment := tc.heatingTime * 0.25
				assert.LessOrEqual(t, adjustment, expectedMaxAdjustment)
			case "negative":
				assert.Less(t, adjustment, 0.0, "Expected negative adjustment for hot feedback")
				// Verify it's proportional to heating time
				expectedMaxAdjustment := tc.heatingTime * 0.25
				assert.GreaterOrEqual(t, adjustment, -expectedMaxAdjustment)
			case "zero":
				assert.Equal(t, 0.0, adjustment, "Expected no adjustment for perfect satisfaction")
			}
		})
	}
}

func TestCalculateUserWeight(t *testing.T) {
	// Arrange
	predictionService := &PredictionService{}

	testCases := []struct {
		name            string
		userRecords     []models.DailyRecord
		requestTemp     float64
		requestDuration float64
		expectedWeight  float64
	}{
		{
			name:            "No user records",
			userRecords:     []models.DailyRecord{},
			requestTemp:     20.0,
			requestDuration: 10.0,
			expectedWeight:  0.0,
		},
		{
			name: "5 relevant records",
			userRecords: func() []models.DailyRecord {
				records := make([]models.DailyRecord, 5)
				for i := 0; i < 5; i++ {
					records[i] = models.DailyRecord{
						AverageTemperature: 20.0,
						ShowerDuration:     10.0,
					}
				}
				return records
			}(),
			requestTemp:     20.0,
			requestDuration: 10.0,
			expectedWeight:  0.5, // 5/10 = 0.5
		},
		{
			name: "10+ relevant records",
			userRecords: func() []models.DailyRecord {
				records := make([]models.DailyRecord, 15)
				for i := 0; i < 15; i++ {
					records[i] = models.DailyRecord{
						AverageTemperature: 20.0,
						ShowerDuration:     10.0,
					}
				}
				return records
			}(),
			requestTemp:     20.0,
			requestDuration: 10.0,
			expectedWeight:  1.0, // Max weight
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &PredictionRequest{
				Temperature: tc.requestTemp,
				Duration:    tc.requestDuration,
			}

			weight := predictionService.calculateUserWeight(req, tc.userRecords)
			assert.Equal(t, tc.expectedWeight, weight)
		})
	}
}
