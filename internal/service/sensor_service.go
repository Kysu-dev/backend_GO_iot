package service

import (
	"smarthome-backend/database/models"
	"smarthome-backend/internal/repository"
	"time"
)

func LogSensorData(temp float64, hum float64, gas int, light int) error {
	data := models.SensorData{
		DeviceID:     2, 
		Timestamp:    time.Now(),
		TemperatureC: temp,
		HumidityPerc: hum,
		GasPPM:       gas,
		LightLux:     light,
	}

	return repository.SaveSensorData(&data)
}

func GetLatestData() (models.SensorData, error) {
	return repository.GetLatestSensorData()
}
