package models

import "time"

// Entity Database
type SensorLight struct {
	LightID   int       `json:"light_id" gorm:"column:light_id"`
	Lux       int       `json:"lux" gorm:"column:lux"`
	Timestamp time.Time `json:"timestamp" gorm:"column:timestamp"`
}

// Request Payload (Input Manual)
type LightRequest struct {
	Lux int `json:"lux" binding:"required"`
}