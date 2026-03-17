package services

import (
	"testing"
	"time"

	"heat-logger/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.DailyRecord{}))
	return db
}

func newTestRecordService(db *gorm.DB) *RecordService {
	return &RecordService{db: db}
}

func createTestRecord(t *testing.T, svc *RecordService, id string) {
	record := &models.DailyRecord{
		ID:                 id,
		UserID:             "test-user",
		Date:               time.Now(),
		ShowerDuration:     10.0,
		AverageTemperature: 20.0,
		HeatingTime:        8.0,
		Satisfaction:       50.0,
	}
	require.NoError(t, svc.CreateRecord(record))
}

func TestDeleteRecord_SoftDeletes(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestRecordService(db)
	createTestRecord(t, svc, "rec-1")

	err := svc.DeleteRecord("rec-1")
	require.NoError(t, err)

	// Should not appear in normal queries
	records, err := svc.GetAllRecords()
	require.NoError(t, err)
	assert.Empty(t, records)

	// Should still exist via Unscoped
	var all []models.DailyRecord
	require.NoError(t, db.Unscoped().Find(&all).Error)
	assert.Len(t, all, 1)
	assert.True(t, all[0].DeletedAt.Valid)
}

func TestGetAllRecords_ExcludesSoftDeleted(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestRecordService(db)
	createTestRecord(t, svc, "rec-1")
	createTestRecord(t, svc, "rec-2")

	require.NoError(t, svc.DeleteRecord("rec-1"))

	records, err := svc.GetAllRecords()
	require.NoError(t, err)
	assert.Len(t, records, 1)
	assert.Equal(t, "rec-2", records[0].ID)
}

func TestRestoreRecord_ClearsDeletedAt(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestRecordService(db)
	createTestRecord(t, svc, "rec-1")

	require.NoError(t, svc.DeleteRecord("rec-1"))

	// Restore
	err := svc.RestoreRecord("rec-1")
	require.NoError(t, err)

	// Should reappear in normal queries
	records, err := svc.GetAllRecords()
	require.NoError(t, err)
	assert.Len(t, records, 1)
	assert.Equal(t, "rec-1", records[0].ID)
}

func TestRestoreRecord_NonExistentID(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestRecordService(db)

	err := svc.RestoreRecord("nonexistent")
	assert.Error(t, err)
	assert.Equal(t, "record not found or not deleted", err.Error())
}

func TestRestoreRecord_NotDeletedRecord(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestRecordService(db)
	createTestRecord(t, svc, "rec-1")

	err := svc.RestoreRecord("rec-1")
	assert.Error(t, err)
	assert.Equal(t, "record not found or not deleted", err.Error())
}

func TestDeleteAllRecords_SoftDeletesAll(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestRecordService(db)
	createTestRecord(t, svc, "rec-1")
	createTestRecord(t, svc, "rec-2")
	createTestRecord(t, svc, "rec-3")

	err := svc.DeleteAllRecords()
	require.NoError(t, err)

	records, err := svc.GetAllRecords()
	require.NoError(t, err)
	assert.Empty(t, records)

	// All should still exist via Unscoped
	var all []models.DailyRecord
	require.NoError(t, db.Unscoped().Find(&all).Error)
	assert.Len(t, all, 3)
}

func TestPredictionQueries_ExcludeSoftDeleted(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestRecordService(db)
	createTestRecord(t, svc, "rec-1")
	createTestRecord(t, svc, "rec-2")

	require.NoError(t, svc.DeleteRecord("rec-1"))

	// GetRecordsForPrediction
	records, err := svc.GetRecordsForPrediction(100)
	require.NoError(t, err)
	assert.Len(t, records, 1)
	assert.Equal(t, "rec-2", records[0].ID)

	// GetRecordsForPredictionByUser
	records, err = svc.GetRecordsForPredictionByUser("test-user", 100)
	require.NoError(t, err)
	assert.Len(t, records, 1)
	assert.Equal(t, "rec-2", records[0].ID)

	// GetGlobalRecordsForPrediction
	records, err = svc.GetGlobalRecordsForPrediction("other-user", 100)
	require.NoError(t, err)
	assert.Len(t, records, 1)
	assert.Equal(t, "rec-2", records[0].ID)
}
