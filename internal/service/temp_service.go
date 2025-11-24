package service

import (
	"smarthome-backend/database/models"
	"smarthome-backend/internal/repository"
	"time"
)

type TempService interface {
	ProcessTemp(temperature float64) error
	GetHistory(limit int) ([]models.SensorTemp, error)
}

type tempService struct {
	repo repository.TempRepository
}

func NewTempService(repo repository.TempRepository) TempService {
	return &tempService{repo: repo}
}

func (s *tempService) ProcessTemp(temperature float64) error {
	data := models.SensorTemp{
		Temperature: temperature,
		Timestamp:   time.Now(),
	}
	return s.repo.Save(&data)
}

func (s *tempService) GetHistory(limit int) ([]models.SensorTemp, error) {
	return s.repo.GetAll(limit)
}
