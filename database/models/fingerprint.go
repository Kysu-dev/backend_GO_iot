package models

import "time"

// FingerprintData represents stored fingerprint template
type FingerprintData struct {
	FingerprintID uint      `gorm:"primaryKey;column:fingerprint_id" json:"fingerprint_id"`
	UserID        uint      `gorm:"not null" json:"user_id"`
	TemplateIndex int       `gorm:"not null" json:"template_index"`
	CreatedAt     time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	User          User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// FingerprintRequest for enrolling new fingerprint
type FingerprintRequest struct {
	UserID        uint `json:"user_id" binding:"required"`
	TemplateIndex int  `json:"template_index" binding:"required"`
}
