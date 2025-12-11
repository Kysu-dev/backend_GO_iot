package repository

import (
	"smarthome-backend/database/models"

	"gorm.io/gorm"
)

type CurtainRepository interface {
	SaveStatus(curtain *models.CurtainStatus) error
	Update(curtain *models.CurtainStatus) error
	GetLatest() (*models.CurtainStatus, error)
}

type curtainRepository struct {
	db *gorm.DB
}

func NewCurtainRepository(db *gorm.DB) CurtainRepository {
	return &curtainRepository{db: db}
}

func (r *curtainRepository) SaveStatus(curtain *models.CurtainStatus) error {
	return r.db.Create(curtain).Error
}

func (r *curtainRepository) Update(curtain *models.CurtainStatus) error {
	// Query 1: Get the existing curtain_id
	var existingID uint
	err := r.db.Model(&models.CurtainStatus{}).
		Select("curtain_id").
		Limit(1).
		Pluck("curtain_id", &existingID).Error

	if err != nil {
		return err
	}

	// Query 2: Update by explicit ID
	return r.db.Model(&models.CurtainStatus{}).
		Where("curtain_id = ?", existingID).
		Updates(map[string]interface{}{
			"status":    curtain.Status,
			"mode":      curtain.Mode,
			"timestamp": curtain.Timestamp,
		}).Error
}

func (r *curtainRepository) GetLatest() (*models.CurtainStatus, error) {
	var curtain models.CurtainStatus
	// Selalu ambil data pertama (karena kita yakin cuma ada 1 data)
	err := r.db.First(&curtain).Error
	return &curtain, err
}
