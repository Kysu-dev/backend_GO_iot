package models

import "time"

// FaceRecognitionLog - Log hasil face recognition dari Python service
type FaceRecognitionLog struct {
	ID         int       `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID     *int      `json:"user_id" gorm:"column:user_id"`
	Name       *string   `json:"name" gorm:"column:name"`
	Confidence float64   `json:"confidence" gorm:"column:confidence"`
	Recognized bool      `json:"recognized" gorm:"column:recognized"`
	Message    string    `json:"message" gorm:"column:message"`
	Source     string    `json:"source" gorm:"column:source"`
	Timestamp  time.Time `json:"timestamp" gorm:"column:timestamp"`
	CreatedAt  time.Time `json:"created_at" gorm:"column:created_at"`
}

func (FaceRecognitionLog) TableName() string {
	return "face_recognition_logs"
}

// FaceAlert - Alert untuk unknown face
type FaceAlert struct {
	ID          int       `json:"id" gorm:"primaryKey;autoIncrement"`
	AlertType   string    `json:"alert_type" gorm:"column:alert_type"`
	ImageBase64 *string   `json:"image_base64" gorm:"column:image_base64;type:longtext"`
	Resolved    bool      `json:"resolved" gorm:"column:resolved;default:false"`
	Timestamp   time.Time `json:"timestamp" gorm:"column:timestamp"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at"`
}

func (FaceAlert) TableName() string {
	return "face_alerts"
}

// FaceRecognitionRequest - Request dari Python service
type FaceRecognitionRequest struct {
	UserID     *int    `json:"user_id"`
	Name       *string `json:"name"`
	Confidence float64 `json:"confidence"`
	Recognized bool    `json:"recognized"`
	Message    string  `json:"message"`
	Timestamp  int64   `json:"timestamp"`
	Source     string  `json:"source"`
}

// FaceAlertRequest - Request alert dari Python service
type FaceAlertRequest struct {
	AlertType   string  `json:"alert_type"`
	ImageBase64 *string `json:"image_base64"`
	Timestamp   int64   `json:"timestamp"`
	Source      string  `json:"source"`
}

// FaceRecognitionEvent - WebSocket broadcast message
type FaceRecognitionEvent struct {
	Event      string  `json:"event"`
	UserID     *int    `json:"user_id,omitempty"`
	Name       *string `json:"name,omitempty"`
	Confidence float64 `json:"confidence"`
	Recognized bool    `json:"recognized"`
	Message    string  `json:"message"`
	Timestamp  int64   `json:"timestamp"`
}
