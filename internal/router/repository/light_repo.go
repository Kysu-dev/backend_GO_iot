package repository

import (
	"smarthome-backend/database/models"

	"gorm.io/gorm"
)

type LightRepository interface {
	Save(data *models.SensorLight) error
	GetAll(limit int) ([]models.SensorLight, error)
}

type lightRepository struct {
	db *gorm.DB
}

func NewLightRepository(db *gorm.DB) LightRepository {
	return &lightRepository{db: db}
}

// Insert data
func (r *lightRepository) Save(data *models.SensorLight) error {
	query := "INSERT INTO sensor_light (lux, timestamp) VALUES (?, ?)"
	return r.db.Exec(query, data.Lux, data.Timestamp).Error
}

// Select all data with limit
func (r *lightRepository) GetAll(limit int) ([]models.SensorLight, error) {
	var results []models.SensorLight
	query := "SELECT * FROM sensor_light ORDER BY timestamp DESC LIMIT ?"
	err := r.db.Raw(query, limit).Scan(&results).Error
	return results, err
}
