package service

import (
	"log"
	"smarthome-backend/database/models"
	"smarthome-backend/internal/repository"
	"time"
)

type CurtainService interface {
	ProcessCurtain(status string, mode string) error
	GetLatest() (*models.CurtainStatus, error)
}

type curtainService struct {
	repo repository.CurtainRepository
}

func NewCurtainService(r repository.CurtainRepository) CurtainService {
	return &curtainService{repo: r}
}

func (s *curtainService) ProcessCurtain(status string, mode string) error {
	if mode == "" { mode = "manual" }

	curtain := &models.CurtainStatus{
		Status:    status, // "open" atau "closed"
		Mode:      mode,
		Timestamp: time.Now(),
	}

	// Panggil Repo SaveStatus (Logika Upsert)
	err := s.repo.SaveStatus(curtain)
	if err != nil {
		log.Printf("❌ Error updating curtain status: %v", err)
		return err
	}

	log.Printf("✅ Curtain updated: %s (Mode: %s)", status, mode)
	return nil
}

func (s *curtainService) GetLatest() (*models.CurtainStatus, error) {
	return s.repo.GetLatest()
}