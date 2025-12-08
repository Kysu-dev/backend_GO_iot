package models

import "time"

type PinCode struct {
	ID           uint      `gorm:"primaryKey;column:id" json:"id"`
	UniversalPin string    `gorm:"type:varchar(6);uniqueIndex;not null;column:universal_pin" json:"universal_pin"`
	SetBy        uint      `gorm:"not null;column:set_by" json:"set_by"`
	CreatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

type PinRequest struct {
	UniversalPin string `json:"universal_pin" binding:"required,min=4,max=6"`
}
