package services

import (
	"testing"
	"time"

	"heat-logger/internal/models"

	"github.com/stretchr/testify/assert"
)

// --- V1 tests ---

func TestV1_ColdStart_SampleCountZero(t *testing.T) {
	mockRS := &MockRecordService{}
	svc := &PredictionService{recordService: mockRS}

	mockRS.On("GetRecordsForPredictionByUser", "new", 50).Return([]models.DailyRecord{}, nil)
	mockRS.On("GetGlobalRecordsForPrediction", "new", 200).Return([]models.DailyRecord{}, nil)

	resp, err := svc.PredictHeatingTime(&PredictionRequest{UserID: "new", Duration: 10, Temperature: 20})
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.SampleCount)
	assert.Nil(t, resp.ConfidenceMin)
	assert.Nil(t, resp.ConfidenceMax)
	assert.Greater(t, resp.HeatingTime, 0.0)
}

func TestV1_SingleMatch_NoRange(t *testing.T) {
	mockRS := &MockRecordService{}
	svc := &PredictionService{recordService: mockRS}

	records := []models.DailyRecord{
		{
			UserID:             "u1",
			Date:               time.Now().AddDate(0, 0, -1),
			ShowerDuration:     10.0,
			AverageTemperature: 20.0,
			HeatingTime:        8.0,
			Satisfaction:       50.0,
		},
	}

	mockRS.On("GetRecordsForPredictionByUser", "u1", 50).Return(records, nil)
	mockRS.On("GetGlobalRecordsForPrediction", "u1", 200).Return([]models.DailyRecord{}, nil)

	resp, err := svc.PredictHeatingTime(&PredictionRequest{UserID: "u1", Duration: 10, Temperature: 20})
	assert.NoError(t, err)
	assert.Equal(t, 1, resp.SampleCount)
	assert.Nil(t, resp.ConfidenceMin)
	assert.Nil(t, resp.ConfidenceMax)
}

func TestV1_MultiMatch_RangePresent(t *testing.T) {
	mockRS := &MockRecordService{}
	svc := &PredictionService{recordService: mockRS}

	records := []models.DailyRecord{
		{UserID: "u1", Date: time.Now().AddDate(0, 0, -1), ShowerDuration: 10, AverageTemperature: 20, HeatingTime: 7.0, Satisfaction: 45},
		{UserID: "u1", Date: time.Now().AddDate(0, 0, -2), ShowerDuration: 10, AverageTemperature: 20, HeatingTime: 9.0, Satisfaction: 50},
		{UserID: "u1", Date: time.Now().AddDate(0, 0, -3), ShowerDuration: 11, AverageTemperature: 21, HeatingTime: 12.0, Satisfaction: 55},
	}

	mockRS.On("GetRecordsForPredictionByUser", "u1", 50).Return(records, nil)
	mockRS.On("GetGlobalRecordsForPrediction", "u1", 200).Return([]models.DailyRecord{}, nil)

	resp, err := svc.PredictHeatingTime(&PredictionRequest{UserID: "u1", Duration: 10, Temperature: 20})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, resp.SampleCount, 2)
	assert.NotNil(t, resp.ConfidenceMin)
	assert.NotNil(t, resp.ConfidenceMax)
	assert.Less(t, *resp.ConfidenceMin, *resp.ConfidenceMax)
}

// --- V2 tests ---

func TestV2_ColdStart_SampleCountZero(t *testing.T) {
	mockRS := &MockRecordService{}
	svc := NewPredictionServiceV2(mockRS, nil)

	mockRS.On("GetRecordsForPredictionByUser", "new", 400).Return([]models.DailyRecord{}, nil)
	mockRS.On("GetGlobalRecordsForPrediction", "new", 1200).Return([]models.DailyRecord{}, nil)

	resp, err := svc.Predict(PredictionRequest{UserID: "new", Duration: 10, Temperature: 20})
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.SampleCount)
	assert.Nil(t, resp.ConfidenceMin)
	assert.Nil(t, resp.ConfidenceMax)
	assert.Greater(t, resp.HeatingTime, 0.0)
}

func TestV2_SingleRecord_NoRange(t *testing.T) {
	mockRS := &MockRecordService{}
	svc := NewPredictionServiceV2(mockRS, nil)

	records := []models.DailyRecord{
		{UserID: "u1", Date: time.Now().AddDate(0, 0, -1), ShowerDuration: 10, AverageTemperature: 20, HeatingTime: 8, Satisfaction: 50},
	}

	mockRS.On("GetRecordsForPredictionByUser", "u1", 400).Return(records, nil)
	mockRS.On("GetGlobalRecordsForPrediction", "u1", 1200).Return([]models.DailyRecord{}, nil)

	resp, err := svc.Predict(PredictionRequest{UserID: "u1", Duration: 10, Temperature: 20})
	assert.NoError(t, err)
	assert.Equal(t, 1, resp.SampleCount)
	assert.Nil(t, resp.ConfidenceMin)
	assert.Nil(t, resp.ConfidenceMax)
}

func TestV2_MultiRecords_RangePresent(t *testing.T) {
	mockRS := &MockRecordService{}
	svc := NewPredictionServiceV2(mockRS, nil)

	records := make([]models.DailyRecord, 6)
	for i := 0; i < 6; i++ {
		records[i] = models.DailyRecord{
			UserID:             "u1",
			Date:               time.Now().AddDate(0, 0, -i-1),
			ShowerDuration:     10 + float64(i%3),
			AverageTemperature: 20 + float64(i%2),
			HeatingTime:        7 + float64(i), // 7, 8, 9, 10, 11, 12
			Satisfaction:       50,
		}
	}

	mockRS.On("GetRecordsForPredictionByUser", "u1", 400).Return(records, nil)
	mockRS.On("GetGlobalRecordsForPrediction", "u1", 1200).Return([]models.DailyRecord{}, nil)

	resp, err := svc.Predict(PredictionRequest{UserID: "u1", Duration: 10, Temperature: 20})
	assert.NoError(t, err)
	assert.Greater(t, resp.SampleCount, 1)
	assert.NotNil(t, resp.ConfidenceMin)
	assert.NotNil(t, resp.ConfidenceMax)
	assert.LessOrEqual(t, *resp.ConfidenceMin, *resp.ConfidenceMax)
}
