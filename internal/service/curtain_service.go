package service

import (
	"log"
	"smarthome-backend/database/models"
	"smarthome-backend/internal/repository"
)

type CurtainService interface {
	ProcessCurtain(position int, mode string) error
	GetLatest() (*models.CurtainStatus, error)
	GetHistory(limit int) ([]models.CurtainStatus, error)
}

type curtainService struct {
	repo repository.CurtainRepository
}

func NewCurtainService(r repository.CurtainRepository) CurtainService {
	return &curtainService{repo: r}
}

func (s *curtainService) ProcessCurtain(position int, mode string) error {
	curtain := &models.CurtainStatus{
		Position: position,
		Mode:     mode,
	}

	err := s.repo.Create(curtain)
	if err != nil {
		log.Printf("❌ Error saving curtain status: %v", err)
		return err
	}

	log.Printf("✅ Curtain status saved: position %d%% (%s)", position, mode)
	return nil
}

func (s *curtainService) GetLatest() (*models.CurtainStatus, error) {
	return s.repo.GetLatest()
}

func (s *curtainService) GetHistory(limit int) ([]models.CurtainStatus, error) {
	return s.repo.GetHistory(limit)
}
