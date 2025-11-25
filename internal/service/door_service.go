package service

import (
	"log"
	"smarthome-backend/database/models"
	"smarthome-backend/internal/repository"
)

type DoorService interface {
	ProcessDoor(status, method string) error
	GetLatest() (*models.DoorStatus, error)
	GetHistory(limit int) ([]models.DoorStatus, error)
}

type doorService struct {
	repo repository.DoorRepository
}

func NewDoorService(r repository.DoorRepository) DoorService {
	return &doorService{repo: r}
}

func (s *doorService) ProcessDoor(status, method string) error {
	door := &models.DoorStatus{
		Status: status,
		Method: method,
	}

	err := s.repo.Create(door)
	if err != nil {
		log.Printf("❌ Error saving door status: %v", err)
		return err
	}

	log.Printf("✅ Door status saved: %s via %s", status, method)
	return nil
}

func (s *doorService) GetLatest() (*models.DoorStatus, error) {
	return s.repo.GetLatest()
}

func (s *doorService) GetHistory(limit int) ([]models.DoorStatus, error) {
	return s.repo.GetHistory(limit)
}
