package repository

import (
	"smarthome-backend/database/models"

	"gorm.io/gorm"
)

type TempRepository interface {
	Save(data *models.SensorTemp) error
	GetAll(limit int) ([]models.SensorTemp, error)
}

type tempRepository struct {
	db *gorm.DB
}

func NewTempRepository(db *gorm.DB) TempRepository {
	return &tempRepository{db: db}
}

// Insert data
func (r *tempRepository) Save(data *models.SensorTemp) error {
	query := "INSERT INTO sensor_temp (temperature, timestamp) VALUES (?, ?)"
	return r.db.Exec(query, data.Temperature, data.Timestamp).Error
}

// Select all data with limit
func (r *tempRepository) GetAll(limit int) ([]models.SensorTemp, error) {
	var results []models.SensorTemp
	query := "SELECT * FROM sensor_temp ORDER BY timestamp DESC LIMIT ?"
	err := r.db.Raw(query, limit).Scan(&results).Error
	return results, err
}
