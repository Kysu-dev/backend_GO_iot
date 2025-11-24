package repository

import (
	"smarthome-backend/config"
	"smarthome-backend/database/models"
)

func SaveSensorData(data *models.SensorData) error {

    sql := "INSERT INTO sensor_data (device_id, timestamp, temperature_c, humidity_perc, gas_ppm, light_lux) VALUES (?, ?, ?, ?, ?, ?)"

    return config.DB.Exec(sql, 
        data.DeviceID, 
        data.Timestamp, 
        data.TemperatureC, 
        data.HumidityPerc, 
        data.GasPPM, 
        data.LightLux,
    ).Error
}

func GetLatestSensorData() (models.SensorData, error) {
	var data models.SensorData
	sql := "SELECT * FROM sensor_data ORDER BY timestamp DESC LIMIT 1"
    err := config.DB.Raw(sql).Scan(&data).Error
	return data, err
}
