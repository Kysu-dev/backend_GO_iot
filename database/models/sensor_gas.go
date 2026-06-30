package models

import "time"

// SensorGas represents gas sensor readings
type SensorGas struct {
	GasID     uint      `gorm:"primaryKey;column:gas_id" json:"gas_id"`
	PPMValue  int       `gorm:"not null" json:"ppm_value"`
	Status    string    `gorm:"type:enum('normal','warning','danger');default:'normal'" json:"status"`
	Timestamp time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"timestamp"`
}

// GasRequest for submitting gas sensor data
type GasRequest struct {
	PPM int `json:"ppm" binding:"required"`
}
