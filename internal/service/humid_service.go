package service

import (
	"smarthome-backend/database/models"
	"smarthome-backend/internal/repository"
	"time"
)

type HumidService interface {
	ProcessHumid(humidity float64) error
	GetHistory(limit int) ([]models.SensorHumid, error)
}

type humidService struct {
	repo repository.HumidRepository
}

func NewHumidService(repo repository.HumidRepository) HumidService {
	return &humidService{repo: repo}
}

func (s *humidService) ProcessHumid(humidity float64) error {
	data := models.SensorHumid{
		Humidity:  humidity,
		Timestamp: time.Now(),
	}
	return s.repo.Save(&data)
}

func (s *humidService) GetHistory(limit int) ([]models.SensorHumid, error) {
	return s.repo.GetAll(limit)
}
