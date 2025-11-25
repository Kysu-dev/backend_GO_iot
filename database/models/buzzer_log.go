package models

import "time"

// BuzzerLog represents buzzer activity log
type BuzzerLog struct {
	BuzzerID  uint      `gorm:"primaryKey;column:buzzer_id" json:"buzzer_id"`
	Status    string    `gorm:"type:enum('on','off')" json:"status"`
	Reason    string    `gorm:"type:varchar(200)" json:"reason,omitempty"`
	Timestamp time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"timestamp"`
}

// BuzzerRequest for controlling buzzer
type BuzzerRequest struct {
	Status string `json:"status" binding:"required,oneof=on off"`
	Reason string `json:"reason"`
}
