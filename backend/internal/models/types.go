package models

import "time"

type Day struct {
	ID                 string    `json:"id"`
	Date               time.Time `json:"date"`
	ShowerDuration     float64   `json:"showerDuration"`
	AverageTemperature float64   `json:"averageTemperature"`
	Satisfaction       int       `json:"satisfaction"`
	HeatingTime        float64   `json:"heatingTime,omitempty"`
}

type HeatingData struct {
	History []Day `json:"history"`
}
