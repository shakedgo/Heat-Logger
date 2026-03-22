package services

import (
	"testing"
	"time"

	"heat-logger/internal/models"

	"github.com/stretchr/testify/assert"
)

func float64Ptr(v float64) *float64 { return &v }

// Helper: create a v2 service with mock records that have sunshine data.
func setupV2WithSunshine(userRecords, globalRecords []models.DailyRecord) *PredictionServiceV2 {
	mock := &MockRecordService{}
	mock.On("GetRecordsForPredictionByUser", "user1", 400).Return(userRecords, nil)
	mock.On("GetGlobalRecordsForPrediction", "user1", 1200).Return(globalRecords, nil)
	return NewPredictionServiceV2(mock, nil)
}

func TestV2_SunshineAffectsPrediction(t *testing.T) {
	// Two records: same duration + temp, but different sunshine → different heating times.
	// Request with low sunshine should favor the low-sunshine record.
	records := []models.DailyRecord{
		{
			UserID:             "user1",
			Date:               time.Now().AddDate(0, 0, -1),
			ShowerDuration:     10,
			AverageTemperature: 20,
			HeatingTime:        30,
			Satisfaction:       50,
			SunshineHours:      float64Ptr(2.0), // cloudy day → needed more heating
		},
		{
			UserID:             "user1",
			Date:               time.Now().AddDate(0, 0, -2),
			ShowerDuration:     10,
			AverageTemperature: 20,
			HeatingTime:        10,
			Satisfaction:       50,
			SunshineHours:      float64Ptr(10.0), // sunny day → less heating
		},
	}
	svc := setupV2WithSunshine(records, []models.DailyRecord{})

	// Request for a cloudy day (2h sunshine)
	cloudyReq := PredictionRequest{UserID: "user1", Duration: 10, Temperature: 20, SunshineHours: float64Ptr(2.0)}
	cloudyResp, err := svc.Predict(cloudyReq)
	assert.NoError(t, err)

	// Request for a sunny day (10h sunshine)
	sunnyReq := PredictionRequest{UserID: "user1", Duration: 10, Temperature: 20, SunshineHours: float64Ptr(10.0)}
	sunnyResp, err := svc.Predict(sunnyReq)
	assert.NoError(t, err)

	// Cloudy day should predict more heating time than sunny day
	assert.Greater(t, cloudyResp.HeatingTime, sunnyResp.HeatingTime,
		"Cloudy day should require more heating than sunny day")
}

func TestV2_NullSunshineFallsBackTo2D(t *testing.T) {
	// Records without sunshine — algorithm should work as before (2D similarity).
	records := []models.DailyRecord{
		{
			UserID:             "user1",
			Date:               time.Now().AddDate(0, 0, -1),
			ShowerDuration:     10,
			AverageTemperature: 20,
			HeatingTime:        20,
			Satisfaction:       50,
			// SunshineHours is nil
		},
	}
	svc := setupV2WithSunshine(records, []models.DailyRecord{})

	// Request also without sunshine
	req := PredictionRequest{UserID: "user1", Duration: 10, Temperature: 20}
	resp, err := svc.Predict(req)
	assert.NoError(t, err)
	assert.NotZero(t, resp.HeatingTime)
}

func TestV2_NullSunshineInRecordSkipsDimension(t *testing.T) {
	// Request has sunshine, but historical record doesn't → sunshine dimension skipped (wSun=1.0).
	records := []models.DailyRecord{
		{
			UserID:             "user1",
			Date:               time.Now().AddDate(0, 0, -1),
			ShowerDuration:     10,
			AverageTemperature: 20,
			HeatingTime:        20,
			Satisfaction:       50,
			SunshineHours:      nil, // old record without sunshine
		},
	}
	svc := setupV2WithSunshine(records, []models.DailyRecord{})

	req := PredictionRequest{UserID: "user1", Duration: 10, Temperature: 20, SunshineHours: float64Ptr(8.0)}
	resp, err := svc.Predict(req)
	assert.NoError(t, err)
	assert.NotZero(t, resp.HeatingTime, "Should still produce a prediction when record lacks sunshine")
}

func TestV2_SigmaSunControlsGaussianWidth(t *testing.T) {
	records := []models.DailyRecord{
		{
			UserID:             "user1",
			Date:               time.Now().AddDate(0, 0, -1),
			ShowerDuration:     10,
			AverageTemperature: 20,
			HeatingTime:        25,
			Satisfaction:       50,
			SunshineHours:      float64Ptr(5.0),
		},
		{
			UserID:             "user1",
			Date:               time.Now().AddDate(0, 0, -2),
			ShowerDuration:     10,
			AverageTemperature: 20,
			HeatingTime:        15,
			Satisfaction:       50,
			SunshineHours:      float64Ptr(12.0),
		},
	}

	mock := &MockRecordService{}
	mock.On("GetRecordsForPredictionByUser", "user1", 400).Return(records, nil)
	mock.On("GetGlobalRecordsForPrediction", "user1", 1200).Return([]models.DailyRecord{}, nil)

	// Tight sigma → more sensitive to sunshine difference
	tightSvc := NewPredictionServiceV2(mock, &PredictionConfigV2{SigmaSun: 1.0})
	// Wide sigma → less sensitive
	wideSvc := NewPredictionServiceV2(mock, &PredictionConfigV2{SigmaSun: 10.0})

	req := PredictionRequest{UserID: "user1", Duration: 10, Temperature: 20, SunshineHours: float64Ptr(5.0)}

	tightResp, _ := tightSvc.Predict(req)
	wideResp, _ := wideSvc.Predict(req)

	// With tight sigma, the nearby record (sunshine=5) dominates → prediction closer to 25
	// With wide sigma, both records contribute more equally → prediction closer to 20
	assert.Greater(t, tightResp.HeatingTime, wideResp.HeatingTime,
		"Tighter SigmaSun should favor the closer-sunshine record more strongly")
}

func TestFreqCellKey_IncludesSunshine(t *testing.T) {
	// Record with sunshine
	r1 := models.DailyRecord{ShowerDuration: 10, AverageTemperature: 20, SunshineHours: float64Ptr(8.0)}
	key1 := freqCellKey(r1)
	assert.Contains(t, key1, "8", "Cell key should include sunshine hours")

	// Record without sunshine
	r2 := models.DailyRecord{ShowerDuration: 10, AverageTemperature: 20}
	key2 := freqCellKey(r2)
	assert.Contains(t, key2, "?", "Cell key should use '?' for null sunshine")

	// Same duration+temp, different sunshine → different keys
	r3 := models.DailyRecord{ShowerDuration: 10, AverageTemperature: 20, SunshineHours: float64Ptr(3.0)}
	key3 := freqCellKey(r3)
	assert.NotEqual(t, key1, key3, "Different sunshine should produce different cell keys")
}

func TestV2_SunshineValidationRange(t *testing.T) {
	// This tests the validation at the handler level conceptually.
	// Sunshine of 0 is valid (cloudy day), sunshine of 16 is max.
	records := []models.DailyRecord{
		{
			UserID:             "user1",
			Date:               time.Now().AddDate(0, 0, -1),
			ShowerDuration:     10,
			AverageTemperature: 20,
			HeatingTime:        20,
			Satisfaction:       50,
			SunshineHours:      float64Ptr(0), // valid: completely cloudy
		},
	}
	svc := setupV2WithSunshine(records, []models.DailyRecord{})

	// Request with 0 sunshine (not null — valid data point)
	req := PredictionRequest{UserID: "user1", Duration: 10, Temperature: 20, SunshineHours: float64Ptr(0)}
	resp, err := svc.Predict(req)
	assert.NoError(t, err)
	assert.NotZero(t, resp.HeatingTime)
}
