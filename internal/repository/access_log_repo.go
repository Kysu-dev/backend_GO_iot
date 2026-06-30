package repository

import (
	"smarthome-backend/database/models"

	"gorm.io/gorm"
)

type AccessLogRepository interface {
	Create(log *models.AccessLog) error
	GetAll(limit int) ([]models.AccessLog, error)
	GetByUserID(userID uint, limit int) ([]models.AccessLog, error)
	GetByStatus(status string, limit int) ([]models.AccessLog, error)
}

type accessLogRepository struct {
	db *gorm.DB
}

func NewAccessLogRepository(db *gorm.DB) AccessLogRepository {
	return &accessLogRepository{db: db}
}

func (r *accessLogRepository) Create(log *models.AccessLog) error {
	return r.db.Create(log).Error
}

func (r *accessLogRepository) GetAll(limit int) ([]models.AccessLog, error) {
	var logs []models.AccessLog
	err := r.db.Preload("User").Order("timestamp DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

func (r *accessLogRepository) GetByUserID(userID uint, limit int) ([]models.AccessLog, error) {
	var logs []models.AccessLog
	err := r.db.Preload("User").Where("user_id = ?", userID).Order("timestamp DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

func (r *accessLogRepository) GetByStatus(status string, limit int) ([]models.AccessLog, error) {
	var logs []models.AccessLog
	err := r.db.Preload("User").Where("status = ?", status).Order("timestamp DESC").Limit(limit).Find(&logs).Error
	return logs, err
}
