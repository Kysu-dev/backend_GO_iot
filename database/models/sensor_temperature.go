package models

import "time"

// SensorTemperature represents temperature sensor readings
type SensorTemperature struct {
	TempID      uint      `gorm:"primaryKey;column:temp_id" json:"temp_id"`
	Temperature float64   `gorm:"type:float;not null" json:"temperature"`
	Timestamp   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"timestamp"`
}

// TempRequest for submitting temperature data
type TempRequest struct {
	Temperature float64 `json:"temperature" binding:"required"`
}
