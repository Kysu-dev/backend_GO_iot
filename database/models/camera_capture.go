package models

import "time"

// CameraCapture represents captured images from ESP32-CAM
type CameraCapture struct {
	CaptureID    uint      `gorm:"primaryKey;column:capture_id" json:"capture_id"`
	ImagePath    string    `gorm:"type:text;not null" json:"image_path"`
	DetectedFace string    `gorm:"type:varchar(100)" json:"detected_face,omitempty"`
	Timestamp    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"timestamp"`
}

// CameraCaptureRequest for saving camera captures
type CameraCaptureRequest struct {
	ImagePath    string `json:"image_path" binding:"required"`
	DetectedFace string `json:"detected_face"`
}
