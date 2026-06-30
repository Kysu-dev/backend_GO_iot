package models

import "time"

// SensorStats represents statistical analysis of sensor data
type SensorStats struct {
	Average      float64 `json:"avg"`
	Min          float64 `json:"min"`
	Max          float64 `json:"max"`
	Median       float64 `json:"median"`
	StdDeviation float64 `json:"std_dev"`
	Count        int     `json:"count"`
}

// SensorStatsResponse represents the complete statistics response
type SensorStatsResponse struct {
	Temperature SensorStats `json:"temperature"`
	Humidity    SensorStats `json:"humidity"`
	TimeRange   string      `json:"time_range"`
	StartTime   time.Time   `json:"start_time"`
	EndTime     time.Time   `json:"end_time"`
}

// HourlyData represents aggregated sensor data per hour
type HourlyData struct {
	Hour        int     `json:"hour"`
	AvgTemp     float64 `json:"avg_temp"`
	AvgHumidity float64 `json:"avg_humidity"`
	MinTemp     float64 `json:"min_temp"`
	MaxTemp     float64 `json:"max_temp"`
	MinHumidity float64 `json:"min_humidity"`
	MaxHumidity float64 `json:"max_humidity"`
	Count       int     `json:"count"`
}

// SensorDataResponse represents paginated sensor data
type SensorDataResponse struct {
	Data       []CombinedSensorData `json:"data"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	TotalPages int                  `json:"total_pages"`
}

// CombinedSensorData combines temperature and humidity readings
type CombinedSensorData struct {
	Timestamp   time.Time `json:"timestamp"`
	Temperature float64   `json:"temperature"`
	Humidity    float64   `json:"humidity"`
}
