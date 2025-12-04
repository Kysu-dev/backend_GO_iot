package repository

import (
	"smarthome-backend/database/models"

	"gorm.io/gorm"
)

type NotificationRepository interface {
	Create(notif *models.Notification) error
	GetAll(limit int) ([]models.Notification, error)
	GetByType(notifType string, limit int) ([]models.Notification, error)
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(notif *models.Notification) error {
	return r.db.Create(notif).Error
}

func (r *notificationRepository) GetAll(limit int) ([]models.Notification, error) {
	var notifs []models.Notification
	err := r.db.Order("timestamp DESC").Limit(limit).Find(&notifs).Error
	return notifs, err
}

func (r *notificationRepository) GetByType(notifType string, limit int) ([]models.Notification, error) {
	var notifs []models.Notification
	err := r.db.Where("type = ?", notifType).Order("timestamp DESC").Limit(limit).Find(&notifs).Error
	return notifs, err
}
