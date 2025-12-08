package models

import "time"

// User represents a system user
type User struct {
	UserID           uint      `gorm:"primaryKey;column:user_id" json:"user_id"`
	Name             string    `gorm:"type:varchar(100);not null" json:"name"`
	Email            string    `gorm:"type:varchar(150);uniqueIndex;not null" json:"email"`
	Password         string    `gorm:"type:varchar(255);not null" json:"-"`
	Role             string    `gorm:"type:enum('admin','member');default:'member'" json:"role"`
	Status           string    `gorm:"type:enum('pending','active','suspended');default:'pending'" json:"status"`
	FaceEncodingPath string    `gorm:"type:text" json:"face_encoding_path,omitempty"`
	CreatedAt        time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt        time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// UserRequest for creating/updating users
type UserRequest struct {
	Name      string `json:"name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	Role      string `json:"role"`
	FaceImage string `json:"face_image"`
}

// LoginRequest for user authentication
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse after successful authentication
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
