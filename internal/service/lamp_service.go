package service

import (
	"log"
	"smarthome-backend/database/models"
	"smarthome-backend/internal/repository"
	"time"
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
		Status:    status,
		Mode:      mode,
		Timestamp: time.Now(),
	}

	// Cek apakah sudah ada data
	existing, err := s.repo.GetLatest()
	if err != nil || existing.LampID == 0 {
		// Belum ada data, INSERT
		err = s.repo.Create(lamp)
		if err != nil {
			log.Printf("Error creating lamp status: %v", err)
			return err
		}
		log.Printf("Lamp status created: %s (%s)", status, mode)
	} else {
		// Sudah ada data, UPDATE
		err = s.repo.Update(lamp)
		if err != nil {
			log.Printf("Error updating lamp status: %v", err)
			return err
		}
		log.Printf("Lamp status updated: %s (%s)", status, mode)
	}

	return nil
}

func (s *lampService) GetLatest() (*models.LampStatus, error) {
	lamp, err := s.repo.GetLatest()

	// --- REVISI: Handling Data Kosong ---
	if err != nil {
		// Kembalikan Default: Mati & Auto
		return &models.LampStatus{
			Status: "off",
			Mode:   "auto",
		}, nil
	}

	return lamp, nil
}

func (s *lampService) GetHistory(limit int) ([]models.LampStatus, error) {
	return s.repo.GetHistory(limit)
}
