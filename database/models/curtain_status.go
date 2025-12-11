package models

import "time"

type CurtainStatus struct {
	CurtainID uint      `gorm:"primaryKey;column:curtain_id" json:"curtain_id"`
	Status    string    `gorm:"type:enum('open','closed')" json:"status"`
	Mode      string    `gorm:"type:enum('auto','manual')" json:"mode"`
	Timestamp time.Time `gorm:"autoUpdateTime" json:"timestamp"`
}

func (CurtainStatus) TableName() string {
	return "curtain_status"
}

type CurtainRequest struct {
	Status string `json:"status" binding:"required,oneof=open closed"`
	Mode   string `json:"mode" binding:"required,oneof=auto manual"`
}
