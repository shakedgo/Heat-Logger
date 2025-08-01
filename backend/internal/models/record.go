package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DailyRecord represents a daily heating record with user feedback
type DailyRecord struct {
	ID                 string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Date               time.Time `json:"date" gorm:"not null"`
	ShowerDuration     float64   `json:"showerDuration" gorm:"not null"`
	AverageTemperature float64   `json:"averageTemperature" gorm:"not null"`
	HeatingTime        float64   `json:"heatingTime" gorm:"not null"`
	Satisfaction       float64   `json:"satisfaction" gorm:"not null"`
	CreatedAt          time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt          time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// BeforeCreate is a GORM hook that generates a UUID before creating a record
func (r *DailyRecord) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}

// TableName specifies the table name for the DailyRecord model
func (DailyRecord) TableName() string {
	return "daily_records"
}
