package repository

import (
	"smarthome-backend/database/models"

	"gorm.io/gorm"
)

type PinRepository interface {
	GetUniversalPin() (*models.PinCode, error)
	Create(pin string, setBy uint) error
	Update(pin string, setBy uint) error
}

type pinRepository struct {
	db *gorm.DB
}

func NewPinRepository(db *gorm.DB) PinRepository {
	return &pinRepository{db: db}
}

func (r *pinRepository) GetUniversalPin() (*models.PinCode, error) {
	var pin models.PinCode
	err := r.db.Order("updated_at desc").First(&pin).Error
	if pin.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &pin, err
}

// Create - Insert PIN pertama kali
func (r *pinRepository) Create(pin string, setBy uint) error {
	newPin := models.PinCode{
		UniversalPin: pin,
		SetBy:        setBy,
	}
	return r.db.Create(&newPin).Error
}

// Update - Update PIN yang sudah ada
func (r *pinRepository) Update(pin string, setBy uint) error {
	// Get latest PIN ID first
	var latestID uint
	err := r.db.Raw("SELECT id FROM pin_codes ORDER BY updated_at DESC LIMIT 1").Scan(&latestID).Error
	if err != nil {
		return err
	}

	// Update using the retrieved ID
	query := "UPDATE pin_codes SET universal_pin = ?, set_by = ?, updated_at = NOW() WHERE id = ?"
	return r.db.Exec(query, pin, setBy, latestID).Error
}
