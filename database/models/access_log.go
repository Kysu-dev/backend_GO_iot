package models

import "time"

// AccessLog represents door access history
type AccessLog struct {
	AccessID  uint      `gorm:"primaryKey;column:access_id" json:"access_id"`
	UserID    *uint     `gorm:"index" json:"user_id,omitempty"`
	Method    string    `gorm:"type:enum('fingerprint','pin','remote','unknown')" json:"method"`
	Status    string    `gorm:"type:enum('success','failed')" json:"status"`
	ImagePath string    `gorm:"type:text" json:"image_path,omitempty"`
	Timestamp time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"timestamp"`
	User      *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// AccessLogRequest for logging access attempts
type AccessLogRequest struct {
	UserID    *uint  `json:"user_id"`
	Method    string `json:"method" binding:"required,oneof=fingerprint pin remote unknown"`
	Status    string `json:"status" binding:"required,oneof=success failed"`
	ImagePath string `json:"image_path"`
}
