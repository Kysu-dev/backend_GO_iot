package models

import "time"

// Entity Database
type SensorTemp struct {
	TempID      int       `json:"temp_id" gorm:"column:temp_id"`
	Temperature float64   `json:"temperature" gorm:"column:temperature"`
	Timestamp   time.Time `json:"timestamp" gorm:"column:timestamp"`
}

// Request Payload (Input Manual)
type TempRequest struct {
	Temperature float64 `json:"temperature" binding:"required"`
}