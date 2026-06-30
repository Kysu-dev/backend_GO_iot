package repository

import (
    "smarthome-backend/database/models"

    "gorm.io/gorm"
)

type LampRepository interface {
    Create(lamp *models.LampStatus) error
    Update(lamp *models.LampStatus) error
    GetLatest() (*models.LampStatus, error)
    GetHistory(limit int) ([]models.LampStatus, error)
}

type lampRepository struct {
    db *gorm.DB
}

func NewLampRepository(db *gorm.DB) LampRepository {
    return &lampRepository{db: db}
}

// INSERT Manual
func (r *lampRepository) Create(lamp *models.LampStatus) error {
    query := "INSERT INTO lamp_status (status, mode, timestamp) VALUES (?, ?, ?)"
    return r.db.Exec(query, lamp.Status, lamp.Mode, lamp.Timestamp).Error
}

// UPDATE Manual (Update data terakhir) - FIXED
func (r *lampRepository) Update(lamp *models.LampStatus) error {
    // Get latest lamp_id first
    var latestID int
    err := r.db.Raw("SELECT lamp_id FROM lamp_status ORDER BY timestamp DESC LIMIT 1").Scan(&latestID).Error
    if err != nil {
        return err
    }

    // Update using the retrieved lamp_id
    query := "UPDATE lamp_status SET status = ?, mode = ?, timestamp = ? WHERE lamp_id = ?"
    return r.db.Exec(query, lamp.Status, lamp.Mode, lamp.Timestamp, latestID).Error
}

// SELECT Manual (Ambil 1 Terakhir)
func (r *lampRepository) GetLatest() (*models.LampStatus, error) {
    var lamp models.LampStatus

    query := "SELECT * FROM lamp_status ORDER BY timestamp DESC LIMIT 1"
    err := r.db.Raw(query).Scan(&lamp).Error

    // Cek jika data kosong (ID 0)
    if lamp.LampID == 0 {
        return nil, gorm.ErrRecordNotFound
    }

    return &lamp, err
}

// SELECT Manual (History)
func (r *lampRepository) GetHistory(limit int) ([]models.LampStatus, error) {
    var lamps []models.LampStatus

    query := "SELECT * FROM lamp_status ORDER BY timestamp DESC LIMIT ?"
    err := r.db.Raw(query, limit).Scan(&lamps).Error

    return lamps, err
}