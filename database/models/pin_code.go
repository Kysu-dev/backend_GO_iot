package models

import "time"

// PinCode represents PIN code for door access
type PinCode struct {
	PinID     uint      `gorm:"primaryKey;column:pin_id" json:"pin_id"`
	PinCode   string    `gorm:"type:varchar(20);not null" json:"pin_code"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}

// PinRequest for creating/updating PIN
type PinRequest struct {
	PinCode string `json:"pin_code" binding:"required,min=4,max=20"`
}
