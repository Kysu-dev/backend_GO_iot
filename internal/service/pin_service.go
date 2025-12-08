package service

import (
	"smarthome-backend/database/models"
	"smarthome-backend/internal/repository"
)

type PinService interface {
	GetUniversalPin() (*models.PinCode, error)
	SetUniversalPin(pin string, setBy uint) error
}

type pinService struct {
	repo repository.PinRepository
}

func NewPinService(r repository.PinRepository) PinService {
	return &pinService{repo: r}
}

func (s *pinService) GetUniversalPin() (*models.PinCode, error) {
	return s.repo.GetUniversalPin()
}

func (s *pinService) SetUniversalPin(pin string, setBy uint) error {
	return s.repo.SetUniversalPin(pin, setBy)
}
