package repository

import (
	"smarthome-backend/database/models"

	"gorm.io/gorm"
)

type HumidRepository interface {
	Save(data *models.SensorHumid) error
	GetAll(limit int) ([]models.SensorHumid, error)
}

type humidRepository struct {
	db *gorm.DB
}

func NewHumidRepository(db *gorm.DB) HumidRepository {
	return &humidRepository{db: db}
}

// Insert data
func (r *humidRepository) Save(data *models.SensorHumid) error {
	query := "INSERT INTO sensor_humid (humidity, timestamp) VALUES (?, ?)"
	return r.db.Exec(query, data.Humidity, data.Timestamp).Error
}

// Select all data with limit
func (r *humidRepository) GetAll(limit int) ([]models.SensorHumid, error) {
	var results []models.SensorHumid
	query := "SELECT * FROM sensor_humid ORDER BY timestamp DESC LIMIT ?"
	err := r.db.Raw(query, limit).Scan(&results).Error
	return results, err
}
