package repository

import (
	"smarthome-backend/database/models"

	"gorm.io/gorm"
)

type HumidRepository interface {
	Save(data *models.SensorHumidity) error
	GetAll(limit int) ([]models.SensorHumidity, error)
}

type humidRepository struct {
	db *gorm.DB
}

func NewHumidRepository(db *gorm.DB) HumidRepository {
	return &humidRepository{db: db}
}

// Insert data
func (r *humidRepository) Save(data *models.SensorHumidity) error {
	query := "INSERT INTO sensor_humidity (humidity, timestamp) VALUES (?, ?)"
	return r.db.Exec(query, data.Humidity, data.Timestamp).Error
}

// Select all data with limit
func (r *humidRepository) GetAll(limit int) ([]models.SensorHumidity, error) {
	var results []models.SensorHumidity
	query := "SELECT * FROM sensor_humidity ORDER BY timestamp DESC LIMIT ?"
	err := r.db.Raw(query, limit).Scan(&results).Error
	return results, err
}
