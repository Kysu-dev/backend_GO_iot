package service

import (
	"log"
	"smarthome-backend/database/models"
	"smarthome-backend/internal/repository"
)

type NotificationService interface {
	Create(req models.NotificationRequest) error
	GetAll(limit int) ([]models.Notification, error)
	GetByType(notifType string, limit int) ([]models.Notification, error)
}

type notificationService struct {
	repo repository.NotificationRepository
}

func NewNotificationService(r repository.NotificationRepository) NotificationService {
	return &notificationService{repo: r}
}

func (s *notificationService) Create(req models.NotificationRequest) error {
	notif := &models.Notification{
		Title:   req.Title,
		Message: req.Message,
		Type:    req.Type,
	}

	err := s.repo.Create(notif)
	if err != nil {
		log.Printf("❌ Error creating notification: %v", err)
		return err
	}

	log.Printf("✅ Notification created: %s (%s)", req.Title, req.Type)
	return nil
}

func (s *notificationService) GetAll(limit int) ([]models.Notification, error) {
	return s.repo.GetAll(limit)
}

func (s *notificationService) GetByType(notifType string, limit int) ([]models.Notification, error) {
	return s.repo.GetByType(notifType, limit)
}
