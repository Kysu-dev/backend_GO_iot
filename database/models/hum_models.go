package models

import "time"

// Entity Database
type SensorHumid struct {
	HumidID   int       `json:"humid_id" gorm:"column:humid_id"`
	Humidity  float64   `json:"humidity" gorm:"column:humidity"`
	Timestamp time.Time `json:"timestamp" gorm:"column:timestamp"`
}

// Request Payload (Input Manual)
type HumidRequest struct {
	Humidity float64 `json:"humidity" binding:"required"`
}