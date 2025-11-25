package models

import "time"

// SensorHumidity represents humidity sensor readings
type SensorHumidity struct {
	HumidID   uint      `gorm:"primaryKey;column:humid_id" json:"humid_id"`
	Humidity  float64   `gorm:"type:float;not null" json:"humidity"`
	Timestamp time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"timestamp"`
}

// HumidRequest for submitting humidity data
type HumidRequest struct {
	Humidity float64 `json:"humidity" binding:"required"`
}
