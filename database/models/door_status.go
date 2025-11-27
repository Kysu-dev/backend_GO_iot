package models

import "time"

// DoorStatus represents current door lock status
type DoorStatus struct {
	DoorID    uint      `gorm:"primaryKey;column:door_id" json:"door_id"`
	Status    string    `gorm:"type:enum('locked','unlocked')" json:"status"`
	Method    string    `gorm:"type:enum('fingerprint','pin','remote','auto')" json:"method"`
	Timestamp time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"timestamp"`
}

func (DoorStatus) TableName() string {
	return "door_status"
}

// DoorRequest for controlling door lock
type DoorRequest struct {
	Status string `json:"status" binding:"required,oneof=locked unlocked"`
	Method string `json:"method" binding:"required,oneof=fingerprint pin remote auto"`
}
