package service

import (
	"log"
	"smarthome-backend/database/models"
	"smarthome-backend/internal/repository"
	"time"
)

type FaceService interface {
	ProcessRecognition(req *models.FaceRecognitionRequest) error
	ProcessAlert(req *models.FaceAlertRequest) error
	GetRecentLogs(limit int) ([]models.FaceRecognitionLog, error)
	GetUnresolvedAlerts() ([]models.FaceAlert, error)
	ResolveAlert(id int) error
}

type faceService struct {
	repo repository.FaceRepository
}

func NewFaceService(repo repository.FaceRepository) FaceService {
	return &faceService{repo: repo}
}

func (s *faceService) ProcessRecognition(req *models.FaceRecognitionRequest) error {
	// 1. Save to database
	data := &models.FaceRecognitionLog{
		UserID:     req.UserID,
		Name:       req.Name,
		Confidence: req.Confidence,
		Recognized: req.Recognized,
		Message:    req.Message,
		Source:     req.Source,
		Timestamp:  time.Unix(req.Timestamp, 0),
	}

	if err := s.repo.SaveRecognitionLog(data); err != nil {
		log.Printf("[Face] Error saving log: %v", err)
		return err
	}

	if req.Recognized {
		if req.Name != nil {
			log.Printf("[Face] Recognized: %s - %.2f%%", *req.Name, req.Confidence*100)
		}
	} else {
		log.Printf("[Face] ⚠️ Unknown face detected")
	}

	return nil
}

func (s *faceService) ProcessAlert(req *models.FaceAlertRequest) error {
	// 1. Save to database
	data := &models.FaceAlert{
		AlertType:   req.AlertType,
		ImageBase64: req.ImageBase64,
		Timestamp:   time.Unix(req.Timestamp, 0),
	}

	if err := s.repo.SaveAlert(data); err != nil {
		log.Printf("[Face] Error saving alert: %v", err)
		return err
	}

	log.Printf("[Face] 🚨 Alert: %s", req.AlertType)

	return nil
}

func (s *faceService) GetRecentLogs(limit int) ([]models.FaceRecognitionLog, error) {
	return s.repo.GetRecentLogs(limit)
}

func (s *faceService) GetUnresolvedAlerts() ([]models.FaceAlert, error) {
	return s.repo.GetUnresolvedAlerts()
}

func (s *faceService) ResolveAlert(id int) error {
	return s.repo.ResolveAlert(id)
}
