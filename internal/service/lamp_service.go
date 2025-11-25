package service

import (
	"log"
	"smarthome-backend/database/models"
	"smarthome-backend/internal/repository"
)

type LampService interface {
	ProcessLamp(status, mode string) error
	GetLatest() (*models.LampStatus, error)
	GetHistory(limit int) ([]models.LampStatus, error)
}

type lampService struct {
	repo repository.LampRepository
}

func NewLampService(r repository.LampRepository) LampService {
	return &lampService{repo: r}
}

func (s *lampService) ProcessLamp(status, mode string) error {
	lamp := &models.LampStatus{
		Status: status,
		Mode:   mode,
	}

	err := s.repo.Create(lamp)
	if err != nil {
		log.Printf("❌ Error saving lamp status: %v", err)
		return err
	}

	log.Printf("✅ Lamp status saved: %s (%s)", status, mode)
	return nil
}

func (s *lampService) GetLatest() (*models.LampStatus, error) {
	return s.repo.GetLatest()
}

func (s *lampService) GetHistory(limit int) ([]models.LampStatus, error) {
	return s.repo.GetHistory(limit)
}
