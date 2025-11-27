package models

import "time"

// LampStatus represents current lamp status
type LampStatus struct {
	LampID    uint      `gorm:"primaryKey;column:lamp_id" json:"lamp_id"`
	Status    string    `gorm:"type:enum('on','off')" json:"status"`
	Mode      string    `gorm:"type:enum('auto','manual')" json:"mode"`
	Timestamp time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"timestamp"`
}

func (LampStatus) TableName() string {
	return "lamp_status"
}

// LampRequest for controlling lamp
type LampRequest struct {
	Status string `json:"status" binding:"required,oneof=on off"`
	Mode   string `json:"mode" binding:"required,oneof=auto manual"`
}
