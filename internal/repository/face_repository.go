package repository

import (
	"smarthome-backend/database/models"
	"time"

	"gorm.io/gorm"
)

type FaceRepository interface {
	SaveRecognitionLog(data *models.FaceRecognitionLog) error
	SaveAlert(data *models.FaceAlert) error
	GetRecentLogs(limit int) ([]models.FaceRecognitionLog, error)
	GetUnresolvedAlerts() ([]models.FaceAlert, error)
	ResolveAlert(id int) error
}

type faceRepository struct {
	db *gorm.DB
}

func NewFaceRepository(db *gorm.DB) FaceRepository {
	return &faceRepository{db: db}
}

func (r *faceRepository) SaveRecognitionLog(data *models.FaceRecognitionLog) error {
	data.CreatedAt = time.Now()
	return r.db.Create(data).Error
}

func (r *faceRepository) SaveAlert(data *models.FaceAlert) error {
	data.CreatedAt = time.Now()
	return r.db.Create(data).Error
}

func (r *faceRepository) GetRecentLogs(limit int) ([]models.FaceRecognitionLog, error) {
	var results []models.FaceRecognitionLog
	err := r.db.Order("timestamp DESC").Limit(limit).Find(&results).Error
	return results, err
}

func (r *faceRepository) GetUnresolvedAlerts() ([]models.FaceAlert, error) {
	var results []models.FaceAlert
	err := r.db.Where("resolved = ?", false).Order("timestamp DESC").Find(&results).Error
	return results, err
}

func (r *faceRepository) ResolveAlert(id int) error {
	return r.db.Model(&models.FaceAlert{}).Where("id = ?", id).Update("resolved", true).Error
}
