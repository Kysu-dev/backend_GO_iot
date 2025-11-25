package models

import "time"

// ==================== USER MODELS ====================

type User struct {
	UserID    uint      `gorm:"primaryKey;column:user_id" json:"user_id"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	Email     string    `gorm:"type:varchar(150);uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"type:varchar(255);not null" json:"-"`
	Role      string    `gorm:"type:enum('admin','member','guest');default:'member'" json:"role"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}

type UserRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// ==================== FINGERPRINT MODELS ====================

type FingerprintData struct {
	FingerprintID uint      `gorm:"primaryKey;column:fingerprint_id" json:"fingerprint_id"`
	UserID        uint      `gorm:"not null" json:"user_id"`
	TemplateIndex int       `gorm:"not null" json:"template_index"`
	CreatedAt     time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	User          User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type FingerprintRequest struct {
	UserID        uint `json:"user_id" binding:"required"`
	TemplateIndex int  `json:"template_index" binding:"required"`
}

// ==================== PIN CODE MODELS ====================

type PinCode struct {
	PinID     uint      `gorm:"primaryKey;column:pin_id" json:"pin_id"`
	PinCode   string    `gorm:"type:varchar(20);not null" json:"pin_code"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}

type PinRequest struct {
	PinCode string `json:"pin_code" binding:"required,min=4,max=20"`
}

// ==================== ACCESS LOG MODELS ====================

type AccessLog struct {
	AccessID  uint      `gorm:"primaryKey;column:access_id" json:"access_id"`
	UserID    *uint     `gorm:"index" json:"user_id,omitempty"`
	Method    string    `gorm:"type:enum('fingerprint','pin','remote','unknown')" json:"method"`
	Status    string    `gorm:"type:enum('success','failed')" json:"status"`
	ImagePath string    `gorm:"type:text" json:"image_path,omitempty"`
	Timestamp time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"timestamp"`
	User      *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type AccessLogRequest struct {
	UserID    *uint  `json:"user_id"`
	Method    string `json:"method" binding:"required,oneof=fingerprint pin remote unknown"`
	Status    string `json:"status" binding:"required,oneof=success failed"`
	ImagePath string `json:"image_path"`
}

// ==================== CAMERA CAPTURE MODELS ====================

type CameraCapture struct {
	CaptureID    uint      `gorm:"primaryKey;column:capture_id" json:"capture_id"`
	ImagePath    string    `gorm:"type:text;not null" json:"image_path"`
	DetectedFace string    `gorm:"type:varchar(100)" json:"detected_face,omitempty"`
	Timestamp    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"timestamp"`
}

type CameraCaptureRequest struct {
	ImagePath    string `json:"image_path" binding:"required"`
	DetectedFace string `json:"detected_face"`
}

// ==================== SENSOR GAS MODELS ====================

type SensorGas struct {
	GasID     uint      `gorm:"primaryKey;column:gas_id" json:"gas_id"`
	PPMValue  int       `gorm:"not null" json:"ppm_value"`
	Status    string    `gorm:"type:enum('normal','warning','danger');default:'normal'" json:"status"`
	Timestamp time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"timestamp"`
}

type GasRequest struct {
	PPM int `json:"ppm" binding:"required"`
}

// ==================== BUZZER LOG MODELS ====================

type BuzzerLog struct {
	BuzzerID  uint      `gorm:"primaryKey;column:buzzer_id" json:"buzzer_id"`
	Status    string    `gorm:"type:enum('on','off')" json:"status"`
	Reason    string    `gorm:"type:varchar(200)" json:"reason,omitempty"`
	Timestamp time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"timestamp"`
}

type BuzzerRequest struct {
	Status string `json:"status" binding:"required,oneof=on off"`
	Reason string `json:"reason"`
}

// ==================== SENSOR TEMPERATURE MODELS ====================

type SensorTemperature struct {
	TempID      uint      `gorm:"primaryKey;column:temp_id" json:"temp_id"`
	Temperature float64   `gorm:"type:float;not null" json:"temperature"`
	Timestamp   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"timestamp"`
}

type TempRequest struct {
	Temperature float64 `json:"temperature" binding:"required"`
}

// ==================== SENSOR HUMIDITY MODELS ====================

type SensorHumidity struct {
	HumidID   uint      `gorm:"primaryKey;column:humid_id" json:"humid_id"`
	Humidity  float64   `gorm:"type:float;not null" json:"humidity"`
	Timestamp time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"timestamp"`
}

type HumidRequest struct {
	Humidity float64 `json:"humidity" binding:"required"`
}

// ==================== SENSOR LIGHT MODELS ====================

type SensorLight struct {
	LightID   uint      `gorm:"primaryKey;column:light_id" json:"light_id"`
	Lux       int       `gorm:"not null" json:"lux"`
	Timestamp time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"timestamp"`
}

type LightRequest struct {
	Lux int `json:"lux" binding:"required"`
}

// ==================== DOOR STATUS MODELS ====================

type DoorStatus struct {
	DoorID    uint      `gorm:"primaryKey;column:door_id" json:"door_id"`
	Status    string    `gorm:"type:enum('locked','unlocked')" json:"status"`
	Method    string    `gorm:"type:enum('fingerprint','pin','remote','auto')" json:"method"`
	Timestamp time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"timestamp"`
}

type DoorRequest struct {
	Status string `json:"status" binding:"required,oneof=locked unlocked"`
	Method string `json:"method" binding:"required,oneof=fingerprint pin remote auto"`
}

// ==================== LAMP STATUS MODELS ====================

type LampStatus struct {
	LampID    uint      `gorm:"primaryKey;column:lamp_id" json:"lamp_id"`
	Status    string    `gorm:"type:enum('on','off')" json:"status"`
	Mode      string    `gorm:"type:enum('auto','manual')" json:"mode"`
	Timestamp time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"timestamp"`
}

type LampRequest struct {
	Status string `json:"status" binding:"required,oneof=on off"`
	Mode   string `json:"mode" binding:"required,oneof=auto manual"`
}

// ==================== CURTAIN STATUS MODELS ====================

type CurtainStatus struct {
	CurtainID uint      `gorm:"primaryKey;column:curtain_id" json:"curtain_id"`
	Position  int       `gorm:"not null" json:"position"`
	Mode      string    `gorm:"type:enum('auto','manual')" json:"mode"`
	Timestamp time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"timestamp"`
}

type CurtainRequest struct {
	Position int    `json:"position" binding:"required,min=0,max=100"`
	Mode     string `json:"mode" binding:"required,oneof=auto manual"`
}

// ==================== NOTIFICATION MODELS ====================

type Notification struct {
	NotifID   uint      `gorm:"primaryKey;column:notif_id" json:"notif_id"`
	Title     string    `gorm:"type:varchar(200)" json:"title"`
	Message   string    `gorm:"type:text" json:"message"`
	Type      string    `gorm:"type:enum('gas','door','system','intruder')" json:"type"`
	Timestamp time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"timestamp"`
}

type NotificationRequest struct {
	Title   string `json:"title" binding:"required"`
	Message string `json:"message" binding:"required"`
	Type    string `json:"type" binding:"required,oneof=gas door system intruder"`
}

// ==================== DEVICE CONTROL MODELS ====================

type DeviceControl struct {
	Device string `json:"device" binding:"required,oneof=door lamp curtain"`
	Action string `json:"action" binding:"required"`
}

// ==================== RESPONSE MODELS ====================

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
