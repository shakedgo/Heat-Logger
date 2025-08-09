package services

import (
	"errors"
	"time"

	"heat-logger/internal/models"
	"heat-logger/pkg/database"

	"gorm.io/gorm"
)

// RecordService handles business logic for daily records
type RecordService struct {
	db *gorm.DB
}

// NewRecordService creates a new record service instance
func NewRecordService() *RecordService {
	return &RecordService{
		db: database.GetDB(),
	}
}

// CreateRecord creates a new daily record
func (s *RecordService) CreateRecord(record *models.DailyRecord) error {
	if record.Date.IsZero() {
		record.Date = time.Now()
	}

	return s.db.Create(record).Error
}

// GetAllRecords retrieves all daily records, ordered by last update descending
func (s *RecordService) GetAllRecords() ([]models.DailyRecord, error) {
	var records []models.DailyRecord
	// Order by UpdatedAt to reflect most recently modified entries first
	err := s.db.Order("updated_at DESC").Find(&records).Error
	return records, err
}

// GetRecordByID retrieves a record by its ID
func (s *RecordService) GetRecordByID(id string) (*models.DailyRecord, error) {
	var record models.DailyRecord
	err := s.db.Where("id = ?", id).First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("record not found")
		}
		return nil, err
	}
	return &record, nil
}

// DeleteRecord deletes a record by its ID
func (s *RecordService) DeleteRecord(id string) error {
	result := s.db.Where("id = ?", id).Delete(&models.DailyRecord{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("record not found")
	}
	return nil
}

// DeleteAllRecords deletes all records
func (s *RecordService) DeleteAllRecords() error {
	return s.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.DailyRecord{}).Error
}

// GetRecordsForPrediction retrieves recent records for ML prediction
func (s *RecordService) GetRecordsForPrediction(limit int) ([]models.DailyRecord, error) {
	var records []models.DailyRecord
	err := s.db.Order("updated_at DESC").Limit(limit).Find(&records).Error
	return records, err
}

// GetRecordsForPredictionByUser retrieves recent records for a specific user for ML prediction
func (s *RecordService) GetRecordsForPredictionByUser(userID string, limit int) ([]models.DailyRecord, error) {
	var records []models.DailyRecord
	err := s.db.Where("user_id = ?", userID).Order("date DESC").Limit(limit).Find(&records).Error
	return records, err
}

// GetGlobalRecordsForPrediction retrieves recent global records (excluding specific user) for ML prediction
func (s *RecordService) GetGlobalRecordsForPrediction(excludeUserID string, limit int) ([]models.DailyRecord, error) {
	var records []models.DailyRecord
	query := s.db.Order("date DESC").Limit(limit)
	if excludeUserID != "" {
		query = query.Where("user_id != ?", excludeUserID)
	}
	err := query.Find(&records).Error
	return records, err
}
