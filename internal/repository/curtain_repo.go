package repository

import (
	"smarthome-backend/database/models"
	"gorm.io/gorm"
)

type CurtainRepository interface {
	SaveStatus(curtain *models.CurtainStatus) error
	GetLatest() (*models.CurtainStatus, error)
}

type curtainRepository struct {
	db *gorm.DB
}

func NewCurtainRepository(db *gorm.DB) CurtainRepository {
	return &curtainRepository{db: db}
}

// --- BAGIAN INI YANG MEMBUAT DIA JADI UPDATE ---
func (r *curtainRepository) SaveStatus(curtain *models.CurtainStatus) error {
	var existing models.CurtainStatus

	// 1. Cek apakah sudah ada data gorden di database? (Ambil 1 data pertama)
	err := r.db.First(&existing).Error

	if err == nil {
		// KONDISI A: DATA SUDAH ADA -> LAKUKAN UPDATE
		// Kita "bajak" ID dari data lama, lalu tempel ke data baru
		curtain.CurtainID = existing.CurtainID
		
		// Simpan perubahan (Ini akan menimpa data lama)
		return r.db.Save(curtain).Error
	}

	// KONDISI B: DATA KOSONG -> LAKUKAN INSERT BARU
	return r.db.Create(curtain).Error
}

func (r *curtainRepository) GetLatest() (*models.CurtainStatus, error) {
	var curtain models.CurtainStatus
	// Selalu ambil data pertama (karena kita yakin cuma ada 1 data)
	err := r.db.First(&curtain).Error
	return &curtain, err
}