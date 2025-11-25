package service

import (
	"log"
	"smarthome-backend/database/models"
	"smarthome-backend/internal/repository"
)

type AccessLogService interface {
	LogAccess(req models.AccessLogRequest) error
	GetAll(limit int) ([]models.AccessLog, error)
	GetByUserID(userID uint, limit int) ([]models.AccessLog, error)
	GetByStatus(status string, limit int) ([]models.AccessLog, error)
}

type accessLogService struct {
	repo repository.AccessLogRepository
}

func NewAccessLogService(r repository.AccessLogRepository) AccessLogService {
	return &accessLogService{repo: r}
}

func (s *accessLogService) LogAccess(req models.AccessLogRequest) error {
	accessLog := &models.AccessLog{
		UserID:    req.UserID,
		Method:    req.Method,
		Status:    req.Status,
		ImagePath: req.ImagePath,
	}

	err := s.repo.Create(accessLog)
	if err != nil {
		log.Printf("❌ Error logging access: %v", err)
		return err
	}

	log.Printf("✅ Access logged: method=%s, status=%s", req.Method, req.Status)
	return nil
}

func (s *accessLogService) GetAll(limit int) ([]models.AccessLog, error) {
	return s.repo.GetAll(limit)
}

func (s *accessLogService) GetByUserID(userID uint, limit int) ([]models.AccessLog, error) {
	return s.repo.GetByUserID(userID, limit)
}

func (s *accessLogService) GetByStatus(status string, limit int) ([]models.AccessLog, error) {
	return s.repo.GetByStatus(status, limit)
}
