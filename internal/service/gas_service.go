package service

import (
	"smarthome-backend/database/models"
	"smarthome-backend/internal/repository"
	"time"
)

type GasService interface {
	// UPDATE: Tambahkan string di return value
	ProcessGas(ppm int) (string, error)
	GetHistory(limit int) ([]models.SensorGas, error)
}

type gasService struct {
	repo repository.GasRepository
}

func NewGasService(repo repository.GasRepository) GasService {
	return &gasService{repo: repo}
}

// UPDATE: Return string status
func (s *gasService) ProcessGas(ppm int) (string, error) {
	status := "normal"
	
	// Logic Penentuan Bahaya
	if ppm > 200 {
		status = "warning"
	}
	if ppm > 500 {
		status = "danger"
	}

	data := models.SensorGas{
		PPMValue:  ppm,
		Status:    status,
		Timestamp: time.Now(),
	}

	return status, s.repo.Save(&data)
}

func (s *gasService) GetHistory(limit int) ([]models.SensorGas, error) {
	return s.repo.GetAll(limit)
}