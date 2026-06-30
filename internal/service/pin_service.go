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
	// Cek apakah sudah ada data
	existing, err := s.repo.GetUniversalPin()
	if err != nil || existing.ID == 0 {
		// Belum ada data, CREATE
		return s.repo.Create(pin, setBy)
	}
	// Sudah ada data, UPDATE
	return s.repo.Update(pin, setBy)
}
