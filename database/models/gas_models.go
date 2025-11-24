package models
import "time"

type SensorGas struct {
	GasID     int       `json:"gas_id" gorm:"column:gas_id"`
	PPMValue  int       `json:"ppm_value" gorm:"column:ppm_value"`
	Status    string    `json:"status" gorm:"column:status"`
	Timestamp time.Time `json:"timestamp" gorm:"column:timestamp"`
}

type GasRequest struct {
	PPM int `json:"ppm_value"`
}