package repository

import (
	"smarthome-backend/database/models"

	"gorm.io/gorm"
)

type DoorRepository interface {
	Create(door *models.DoorStatus) error
	GetLatest() (*models.DoorStatus, error)
	GetHistory(limit int) ([]models.DoorStatus, error)
}

type doorRepository struct {
	db *gorm.DB
}

func NewDoorRepository(db *gorm.DB) DoorRepository {
	return &doorRepository{db: db}
}

func (r *doorRepository) Create(door *models.DoorStatus) error {
	return r.db.Create(door).Error
}

func (r *doorRepository) GetLatest() (*models.DoorStatus, error) {
	var door models.DoorStatus
	err := r.db.Order("timestamp DESC").First(&door).Error
	return &door, err
}

func (r *doorRepository) GetHistory(limit int) ([]models.DoorStatus, error) {
	var doors []models.DoorStatus
	err := r.db.Order("timestamp DESC").Limit(limit).Find(&doors).Error
	return doors, err
}
