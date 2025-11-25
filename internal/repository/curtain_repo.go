package repository

import (
	"smarthome-backend/database/models"

	"gorm.io/gorm"
)

type CurtainRepository interface {
	Create(curtain *models.CurtainStatus) error
	GetLatest() (*models.CurtainStatus, error)
	GetHistory(limit int) ([]models.CurtainStatus, error)
}

type curtainRepository struct {
	db *gorm.DB
}

func NewCurtainRepository(db *gorm.DB) CurtainRepository {
	return &curtainRepository{db: db}
}

func (r *curtainRepository) Create(curtain *models.CurtainStatus) error {
	return r.db.Create(curtain).Error
}

func (r *curtainRepository) GetLatest() (*models.CurtainStatus, error) {
	var curtain models.CurtainStatus
	err := r.db.Order("timestamp DESC").First(&curtain).Error
	return &curtain, err
}

func (r *curtainRepository) GetHistory(limit int) ([]models.CurtainStatus, error) {
	var curtains []models.CurtainStatus
	err := r.db.Order("timestamp DESC").Limit(limit).Find(&curtains).Error
	return curtains, err
}
