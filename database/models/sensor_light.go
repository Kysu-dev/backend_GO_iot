package models

import "time"

// SensorLight represents light sensor readings
type SensorLight struct {
	LightID   uint      `gorm:"primaryKey;column:light_id" json:"light_id"`
	Lux       int       `gorm:"not null" json:"lux"`
	Timestamp time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"timestamp"`
}

// LightRequest for submitting light sensor data
type LightRequest struct {
	Lux int `json:"lux" binding:"required"`
}
