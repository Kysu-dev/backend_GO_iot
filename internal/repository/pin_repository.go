package repository

import (
	"smarthome-backend/database/models"

	"gorm.io/gorm"
)

type PinRepository interface {
	GetUniversalPin() (*models.PinCode, error)
	SetUniversalPin(pin string, setBy uint) error
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
	return &pin, err
}

func (r *pinRepository) SetUniversalPin(pin string, setBy uint) error {
	var pinCode models.PinCode
	err := r.db.Order("updated_at desc").First(&pinCode).Error
	if err == nil {
		pinCode.UniversalPin = pin
		pinCode.SetBy = setBy
		return r.db.Save(&pinCode).Error
	}
	newPin := models.PinCode{
		UniversalPin: pin,
		SetBy:        setBy,
	}
	return r.db.Create(&newPin).Error
}
