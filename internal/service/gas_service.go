package service

import (
	"smarthome-backend/database/models"
	"smarthome-backend/internal/repository"
	"time"
)

type GasService interface {
	ProcessGas(ppm int) error
	GetHistory(limit int) ([]models.SensorGas, error)
}

type gasService struct {
	repo repository.GasRepository
}

func NewGasService(repo repository.GasRepository) GasService {
	return &gasService{repo: repo}
}

func (s *gasService) ProcessGas(ppm int) error {
	status := "normal"
	if ppm > 500 { status = "warning" }
	if ppm > 1000 { status = "danger" }

	data := models.SensorGas{
		PPMValue:  ppm,
		Status:    status,
		Timestamp: time.Now(),
	}
	return s.repo.Save(&data)
}

func (s *gasService) GetHistory(limit int) ([]models.SensorGas, error) {
	return s.repo.GetAll(limit)
}