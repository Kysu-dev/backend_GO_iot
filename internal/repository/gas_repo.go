package repository

import (
	"smarthome-backend/database/models" // <--- Import dari folder database
	"gorm.io/gorm"
)

type GasRepository interface {
	Save(data *models.SensorGas) error
	GetAll(limit int) ([]models.SensorGas, error)
}

type gasRepository struct {
	db *gorm.DB
}

func NewGasRepository(db *gorm.DB) GasRepository {
	return &gasRepository{db: db}
}

// Insert data
func (r *gasRepository) Save(data *models.SensorGas) error {
	query := "INSERT INTO sensor_gas (ppm_value, status, timestamp) VALUES (?, ?, ?)"
	return r.db.Exec(query, data.PPMValue, data.Status, data.Timestamp).Error
}

// Select all data with limit
func (r *gasRepository) GetAll(limit int) ([]models.SensorGas, error) {
	var results []models.SensorGas
	query := "SELECT * FROM sensor_gas ORDER BY timestamp DESC LIMIT ?"
	err := r.db.Raw(query, limit).Scan(&results).Error
	return results, err
}