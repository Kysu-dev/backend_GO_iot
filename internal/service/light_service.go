package service

import (
	"smarthome-backend/database/models"
	"smarthome-backend/internal/repository"
	"time"
)

type LightService interface {
	ProcessLight(lux int) error
	GetHistory(limit int) ([]models.SensorLight, error)
}

type lightService struct {
	repo repository.LightRepository
}

func NewLightService(repo repository.LightRepository) LightService {
	return &lightService{repo: repo}
}

func (s *lightService) ProcessLight(lux int) error {
	data := models.SensorLight{
		Lux:       lux,
		Timestamp: time.Now(),
	}
	return s.repo.Save(&data)
}

func (s *lightService) GetHistory(limit int) ([]models.SensorLight, error) {
	return s.repo.GetAll(limit)
}
