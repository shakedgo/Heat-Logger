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

	// Use Satisfaction=30 (not 50) to exercise the kNN path rather than the anchor branch
	records := []models.DailyRecord{
		{UserID: "u1", Date: time.Now().AddDate(0, 0, -1), ShowerDuration: 10, AverageTemperature: 20, HeatingTime: 8, Satisfaction: 30},
	}

	mockRS.On("GetRecordsForPredictionByUser", "u1", 400).Return(records, nil)
	mockRS.On("GetGlobalRecordsForPrediction", "u1", 1200).Return([]models.DailyRecord{}, nil)

	resp, err := svc.Predict(PredictionRequest{UserID: "u1", Duration: 10, Temperature: 20})
	assert.NoError(t, err)
	assert.Equal(t, 1, resp.SampleCount)
	assert.Nil(t, resp.ConfidenceMin)
	assert.Nil(t, resp.ConfidenceMax)
}

func TestV2_ColdFeedback_PredictionHigherThanHeatingTime(t *testing.T) {
	mockRS := &MockRecordService{}
	svc := NewPredictionServiceV2(mockRS, &PredictionConfigV2{
		StepCapFraction:     0.99,
		RecencyHalfLifeDays: 365,
		MinK:                1,
	})

	records := make([]models.DailyRecord, 6)
	for i := 0; i < 6; i++ {
		records[i] = models.DailyRecord{
			UserID:             "u1",
			Date:               time.Now().AddDate(0, 0, -1),
			ShowerDuration:     10.0,
			AverageTemperature: 20.0,
			HeatingTime:        15.0,
			Satisfaction:       10.0,
		}
	}

	mockRS.On("GetRecordsForPredictionByUser", "u1", 400).Return(records, nil)
	mockRS.On("GetGlobalRecordsForPrediction", "u1", 1200).Return([]models.DailyRecord{}, nil)

	resp, err := svc.Predict(PredictionRequest{UserID: "u1", Duration: 10, Temperature: 20})
	assert.NoError(t, err)
	assert.Greater(t, resp.HeatingTime, 15.0)
}

func TestV2_HotFeedback_PredictionLowerThanHeatingTime(t *testing.T) {
	mockRS := &MockRecordService{}
	svc := NewPredictionServiceV2(mockRS, &PredictionConfigV2{
		StepCapFraction:     0.99,
		RecencyHalfLifeDays: 365,
		MinK:                1,
	})

	records := make([]models.DailyRecord, 6)
	for i := 0; i < 6; i++ {
		records[i] = models.DailyRecord{
			UserID:             "u1",
			Date:               time.Now().AddDate(0, 0, -1),
			ShowerDuration:     10.0,
			AverageTemperature: 20.0,
			HeatingTime:        20.0,
			Satisfaction:       90.0,
		}
	}

	mockRS.On("GetRecordsForPredictionByUser", "u1", 400).Return(records, nil)
	mockRS.On("GetGlobalRecordsForPrediction", "u1", 1200).Return([]models.DailyRecord{}, nil)

	resp, err := svc.Predict(PredictionRequest{UserID: "u1", Duration: 10, Temperature: 20})
	assert.NoError(t, err)
	assert.Less(t, resp.HeatingTime, 20.0)
}

func TestV2_UserBoost_IgnoresGlobalOutlier(t *testing.T) {
	mockRS := &MockRecordService{}
	svc := NewPredictionServiceV2(mockRS, &PredictionConfigV2{
		StepCapFraction:     0.99,
		RecencyHalfLifeDays: 365,
		MinK:                1,
		UserBoost:           5.0,
	})

	userRecords := make([]models.DailyRecord, 12)
	for i := 0; i < 12; i++ {
		userRecords[i] = models.DailyRecord{
			UserID:             "u1",
			Date:               time.Now().AddDate(0, 0, -1),
			ShowerDuration:     10.0,
			AverageTemperature: 20.0,
			HeatingTime:        10.0,
			Satisfaction:       50.0,
		}
	}
	globalRecords := []models.DailyRecord{
		{
			UserID:             "other",
			Date:               time.Now().AddDate(0, 0, -1),
			ShowerDuration:     10.0,
			AverageTemperature: 20.0,
			HeatingTime:        60.0,
			Satisfaction:       50.0,
		},
	}

	mockRS.On("GetRecordsForPredictionByUser", "u1", 400).Return(userRecords, nil)
	mockRS.On("GetGlobalRecordsForPrediction", "u1", 1200).Return(globalRecords, nil)

	resp, err := svc.Predict(PredictionRequest{UserID: "u1", Duration: 10, Temperature: 20})
	assert.NoError(t, err)
	assert.Less(t, resp.HeatingTime, 25.0)
}

func TestV2_NeverCold_ColdStart_ReturnsCeilDefault(t *testing.T) {
	mockRS := &MockRecordService{}
	svc := NewPredictionServiceV2(mockRS, &PredictionConfigV2{
		NeverCold:  true,
		MinMinutes: 5,
		MaxMinutes: 120,
	})

	mockRS.On("GetRecordsForPredictionByUser", "cold-start", 400).Return([]models.DailyRecord{}, nil)
	mockRS.On("GetGlobalRecordsForPrediction", "cold-start", 1200).Return([]models.DailyRecord{}, nil)

	resp, err := svc.Predict(PredictionRequest{UserID: "cold-start", Duration: 10, Temperature: 20})
	assert.NoError(t, err)
	assert.Equal(t, 30.0, resp.HeatingTime)
}

// TG1: near-perfect records (sat≈50) should drive the prediction, not be zeroed out.
// If AnchorBoost were 0 (old broken default) all their weights would be killed and
// weightedMeanTargets would fall back to 30. With the fix they get boosted toward 8.
func TestV2_AnchorBoost_NearPerfectRecordsDrivePrediction(t *testing.T) {
	mockRS := &MockRecordService{}
	svc := NewPredictionServiceV2(mockRS, &PredictionConfigV2{
		RecencyHalfLifeDays: 365,
		StepCapFraction:     0.99,
		MinK:                1,
	})

	records := make([]models.DailyRecord, 6)
	for i := 0; i < 6; i++ {
		records[i] = models.DailyRecord{
			UserID: "u1", Date: time.Now().AddDate(0, 0, -i-1),
			ShowerDuration: 10, AverageTemperature: 20,
			HeatingTime: 8, Satisfaction: 50,
		}
	}

	mockRS.On("GetRecordsForPredictionByUser", "u1", 400).Return(records, nil)
	mockRS.On("GetGlobalRecordsForPrediction", "u1", 1200).Return([]models.DailyRecord{}, nil)

	resp, err := svc.Predict(PredictionRequest{UserID: "u1", Duration: 10, Temperature: 20})
	assert.NoError(t, err)
	assert.Less(t, resp.HeatingTime, 20.0,
		"near-perfect records should drive the prediction toward 8, not fall back to 30")
}

// TG2: exercise all default config values end-to-end with a realistic mixed dataset.
func TestV2_DefaultConfig_EndToEnd(t *testing.T) {
	mockRS := &MockRecordService{}
	svc := NewPredictionServiceV2(mockRS, nil)

	userRecords := []models.DailyRecord{
		{UserID: "u1", Date: time.Now().AddDate(0, 0, -1), ShowerDuration: 10, AverageTemperature: 20, HeatingTime: 12, Satisfaction: 40},
		{UserID: "u1", Date: time.Now().AddDate(0, 0, -2), ShowerDuration: 10, AverageTemperature: 20, HeatingTime: 14, Satisfaction: 50},
		{UserID: "u1", Date: time.Now().AddDate(0, 0, -3), ShowerDuration: 10, AverageTemperature: 20, HeatingTime: 13, Satisfaction: 55},
		{UserID: "u1", Date: time.Now().AddDate(0, 0, -4), ShowerDuration: 10, AverageTemperature: 21, HeatingTime: 11, Satisfaction: 45},
		{UserID: "u1", Date: time.Now().AddDate(0, 0, -5), ShowerDuration: 11, AverageTemperature: 20, HeatingTime: 15, Satisfaction: 50},
		{UserID: "u1", Date: time.Now().AddDate(0, 0, -6), ShowerDuration: 10, AverageTemperature: 20, HeatingTime: 13, Satisfaction: 48},
	}
	globalRecords := []models.DailyRecord{
		{UserID: "other", Date: time.Now().AddDate(0, 0, -1), ShowerDuration: 10, AverageTemperature: 20, HeatingTime: 13, Satisfaction: 50},
	}

	mockRS.On("GetRecordsForPredictionByUser", "u1", 400).Return(userRecords, nil)
	mockRS.On("GetGlobalRecordsForPrediction", "u1", 1200).Return(globalRecords, nil)

	resp, err := svc.Predict(PredictionRequest{UserID: "u1", Duration: 10, Temperature: 20})
	assert.NoError(t, err)
	assert.Greater(t, resp.SampleCount, 0)
	assert.GreaterOrEqual(t, resp.HeatingTime, 5.0)  // MinMinutes default
	assert.LessOrEqual(t, resp.HeatingTime, 120.0)   // MaxMinutes default
}

// TG3: NeverCold must also suppress floor-rounding on the normal (non-empty) prediction path.
func TestV2_NeverCold_NormalPath_NotLowerThanDefault(t *testing.T) {
	records := make([]models.DailyRecord, 6)
	for i := 0; i < 6; i++ {
		records[i] = models.DailyRecord{
			UserID: "u1", Date: time.Now().AddDate(0, 0, -i-1),
			ShowerDuration: 10, AverageTemperature: 20,
			HeatingTime: 15, Satisfaction: 70, // recently hot — triggers floor path in smartRound
		}
	}
	req := PredictionRequest{UserID: "u1", Duration: 10, Temperature: 20}

	mockRS1 := &MockRecordService{}
	mockRS1.On("GetRecordsForPredictionByUser", "u1", 400).Return(records, nil)
	mockRS1.On("GetGlobalRecordsForPrediction", "u1", 1200).Return([]models.DailyRecord{}, nil)
	svcNormal := NewPredictionServiceV2(mockRS1, &PredictionConfigV2{
		NeverCold: false, StepCapFraction: 0.99, RecencyHalfLifeDays: 365, MinK: 1,
	})
	respNormal, err := svcNormal.Predict(req)
	assert.NoError(t, err)

	mockRS2 := &MockRecordService{}
	mockRS2.On("GetRecordsForPredictionByUser", "u1", 400).Return(records, nil)
	mockRS2.On("GetGlobalRecordsForPrediction", "u1", 1200).Return([]models.DailyRecord{}, nil)
	svcNeverCold := NewPredictionServiceV2(mockRS2, &PredictionConfigV2{
		NeverCold: true, StepCapFraction: 0.99, RecencyHalfLifeDays: 365, MinK: 1,
	})
	respNeverCold, err := svcNeverCold.Predict(req)
	assert.NoError(t, err)

	assert.GreaterOrEqual(t, respNeverCold.HeatingTime, respNormal.HeatingTime,
		"NeverCold=true should never produce a lower prediction than NeverCold=false")
}

// TG4: the prediction must fall within [ConfidenceMin, ConfidenceMax].
func TestV2_ConfidenceRange_ContainsPrediction(t *testing.T) {
	mockRS := &MockRecordService{}
	svc := NewPredictionServiceV2(mockRS, &PredictionConfigV2{
		StepCapFraction:     0.99,
		RecencyHalfLifeDays: 365,
		MinK:                1,
		AnchorBlend:         0, // disable anchor blending so estimate stays within kNN range
	})

	records := []models.DailyRecord{
		{UserID: "u1", Date: time.Now().AddDate(0, 0, -1), ShowerDuration: 10, AverageTemperature: 20, HeatingTime: 10, Satisfaction: 50},
		{UserID: "u1", Date: time.Now().AddDate(0, 0, -2), ShowerDuration: 10, AverageTemperature: 20, HeatingTime: 12, Satisfaction: 50},
		{UserID: "u1", Date: time.Now().AddDate(0, 0, -3), ShowerDuration: 10, AverageTemperature: 20, HeatingTime: 11, Satisfaction: 50},
		{UserID: "u1", Date: time.Now().AddDate(0, 0, -4), ShowerDuration: 10, AverageTemperature: 20, HeatingTime: 13, Satisfaction: 50},
		{UserID: "u1", Date: time.Now().AddDate(0, 0, -5), ShowerDuration: 10, AverageTemperature: 20, HeatingTime:  9, Satisfaction: 50},
		{UserID: "u1", Date: time.Now().AddDate(0, 0, -6), ShowerDuration: 10, AverageTemperature: 20, HeatingTime: 14, Satisfaction: 50},
	}
	mockRS.On("GetRecordsForPredictionByUser", "u1", 400).Return(records, nil)
	mockRS.On("GetGlobalRecordsForPrediction", "u1", 1200).Return([]models.DailyRecord{}, nil)

	resp, err := svc.Predict(PredictionRequest{UserID: "u1", Duration: 10, Temperature: 20})
	assert.NoError(t, err)
	assert.NotNil(t, resp.ConfidenceMin)
	assert.NotNil(t, resp.ConfidenceMax)
	assert.LessOrEqual(t, *resp.ConfidenceMin, resp.HeatingTime,
		"prediction should be >= ConfidenceMin")
	assert.LessOrEqual(t, resp.HeatingTime, *resp.ConfidenceMax,
		"prediction should be <= ConfidenceMax")
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
