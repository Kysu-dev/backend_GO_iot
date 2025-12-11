package repository

import (
	"smarthome-backend/database/models"

	"gorm.io/gorm"
)

type DoorRepository interface {
	Create(door *models.DoorStatus) error
	Update(door *models.DoorStatus) error
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

// Update - Update door status yang sudah ada
func (r *doorRepository) Update(door *models.DoorStatus) error {
	// Get latest door_id first
	var latestID uint
	err := r.db.Raw("SELECT door_id FROM door_status ORDER BY timestamp DESC LIMIT 1").Scan(&latestID).Error
	if err != nil {
		return err
	}

	// Update using the retrieved door_id
	query := "UPDATE door_status SET status = ?, method = ?, timestamp = NOW() WHERE door_id = ?"
	return r.db.Exec(query, door.Status, door.Method, latestID).Error
}

func (r *doorRepository) GetLatest() (*models.DoorStatus, error) {
	var door models.DoorStatus
	err := r.db.Order("timestamp DESC").First(&door).Error
	if door.DoorID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &door, err
}

func (r *doorRepository) GetHistory(limit int) ([]models.DoorStatus, error) {
	var doors []models.DoorStatus
	err := r.db.Order("timestamp DESC").Limit(limit).Find(&doors).Error
	return doors, err
}
