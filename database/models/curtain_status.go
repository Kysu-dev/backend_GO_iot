package models

import "time"

// CurtainStatus represents current curtain position
type CurtainStatus struct {
	CurtainID uint      `gorm:"primaryKey;column:curtain_id" json:"curtain_id"`
	Position  int       `gorm:"not null" json:"position"`
	Mode      string    `gorm:"type:enum('auto','manual')" json:"mode"`
	Timestamp time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"timestamp"`
}

func (CurtainStatus) TableName() string {
	return "curtain_status"
}	

// CurtainRequest for controlling curtain position
type CurtainRequest struct {
	Position int    `json:"position" binding:"required,min=0,max=100"`
	Mode     string `json:"mode" binding:"required,oneof=auto manual"`
}
