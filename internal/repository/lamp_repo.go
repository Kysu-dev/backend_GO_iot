package repository

import (
	"smarthome-backend/database/models"

	"gorm.io/gorm"
)

type LampRepository interface {
	Create(lamp *models.LampStatus) error
	GetLatest() (*models.LampStatus, error)
	GetHistory(limit int) ([]models.LampStatus, error)
}

type lampRepository struct {
	db *gorm.DB
}


func NewLampRepository(db *gorm.DB) LampRepository {
	return &lampRepository{db: db}
}

func (r *lampRepository) Create(lamp *models.LampStatus) error {
	return r.db.Create(lamp).Error
}

func (r *lampRepository) GetLatest() (*models.LampStatus, error) {
	var lamp models.LampStatus
	err := r.db.Order("timestamp DESC").First(&lamp).Error
	return &lamp, err
}

func (r *lampRepository) GetHistory(limit int) ([]models.LampStatus, error) {
	var lamps []models.LampStatus
	err := r.db.Order("timestamp DESC").Limit(limit).Find(&lamps).Error
	return lamps, err
}
