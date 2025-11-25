package models

import "time"

// Notification represents system notifications
type Notification struct {
	NotifID   uint      `gorm:"primaryKey;column:notif_id" json:"notif_id"`
	Title     string    `gorm:"type:varchar(200)" json:"title"`
	Message   string    `gorm:"type:text" json:"message"`
	Type      string    `gorm:"type:enum('gas','door','system','intruder')" json:"type"`
	Timestamp time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"timestamp"`
}

// NotificationRequest for creating notifications
type NotificationRequest struct {
	Title   string `json:"title" binding:"required"`
	Message string `json:"message" binding:"required"`
	Type    string `json:"type" binding:"required,oneof=gas door system intruder"`
}
