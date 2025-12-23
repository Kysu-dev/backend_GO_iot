package service

import (
    "errors"
    "math"
    "smarthome-backend/database/models"
    "sort"
    "time"

    "gorm.io/gorm"
)

type SensorAnalyticsService interface {
    GetStatistics(timeRange string) (*models.SensorStatsResponse, error)
    GetPaginatedData(timeRange string, page, pageSize int) (*models.SensorDataResponse, error)
    GetHourlyData(timeRange string) ([]models.HourlyData, error)
}

type sensorAnalyticsService struct {
    db *gorm.DB
}

func NewSensorAnalyticsService(db *gorm.DB) SensorAnalyticsService {
    return &sensorAnalyticsService{db: db}
}

// parseTimeRange converts time range string to start/end times in UTC
func parseTimeRange(timeRange string) (time.Time, time.Time) {
    now := time.Now().UTC()
    var start time.Time

    switch timeRange {
    case "1h":
        start = now.Add(-1 * time.Hour)
    case "6h":
        start = now.Add(-6 * time.Hour)
    case "24h":
        start = now.Add(-24 * time.Hour)
    case "7d":
        start = now.Add(-7 * 24 * time.Hour)
    default:
        start = now.Add(-24 * time.Hour) // default to 24h
    }

    return start, now
}

// calculateStats computes statistics for a slice of float64 values
func calculateStats(values []float64) models.SensorStats {
    if len(values) == 0 {
        return models.SensorStats{}
    }

    sorted := make([]float64, len(values))
    copy(sorted, values)
    sort.Float64s(sorted)

    var sum float64
    for _, v := range values {
        sum += v
    }
    avg := sum / float64(len(values))

    var variance float64
    for _, v := range values {
        variance += math.Pow(v-avg, 2)
    }
    stdDev := math.Sqrt(variance / float64(len(values)))

    var median float64
    mid := len(sorted) / 2
    if len(sorted)%2 == 0 {
        median = (sorted[mid-1] + sorted[mid]) / 2
    } else {
        median = sorted[mid]
    }

    return models.SensorStats{
        Average:      math.Round(avg*100) / 100,
        Min:          sorted[0],
        Max:          sorted[len(sorted)-1],
        Median:       math.Round(median*100) / 100,
        StdDeviation: math.Round(stdDev*100) / 100,
        Count:        len(values),
    }
}

func (s *sensorAnalyticsService) GetStatistics(timeRange string) (*models.SensorStatsResponse, error) {
    start, end := parseTimeRange(timeRange)
    startUTC, endUTC := start.UTC(), end.UTC()

    var tempData []models.SensorTemperature
    if err := s.db.Where("timestamp BETWEEN ? AND ?", startUTC, endUTC).
        Order("timestamp ASC").
        Find(&tempData).Error; err != nil {
        return nil, err
    }

    var humidData []models.SensorHumidity
    if err := s.db.Where("timestamp BETWEEN ? AND ?", startUTC, endUTC).
        Order("timestamp ASC").
        Find(&humidData).Error; err != nil {
        return nil, err
    }

    tempValues := make([]float64, len(tempData))
    for i, t := range tempData {
        tempValues[i] = t.Temperature
    }

    humidValues := make([]float64, len(humidData))
    for i, h := range humidData {
        humidValues[i] = h.Humidity
    }

    tempStats := calculateStats(tempValues)
    humidStats := calculateStats(humidValues)

    return &models.SensorStatsResponse{
        Temperature: tempStats,
        Humidity:    humidStats,
        TimeRange:   timeRange,
        StartTime:   startUTC,
        EndTime:     endUTC,
    }, nil
}

func (s *sensorAnalyticsService) GetPaginatedData(timeRange string, page, pageSize int) (*models.SensorDataResponse, error) {
    start, end := parseTimeRange(timeRange)
    startUTC, endUTC := start.UTC(), end.UTC()

    var tempData []models.SensorTemperature
    var totalTemp int64

    offset := (page - 1) * pageSize

    if err := s.db.Model(&models.SensorTemperature{}).
        Where("timestamp BETWEEN ? AND ?", startUTC, endUTC).
        Count(&totalTemp).Error; err != nil {
        return nil, err
    }

    if err := s.db.Where("timestamp BETWEEN ? AND ?", startUTC, endUTC).
        Order("timestamp DESC").
        Limit(pageSize).
        Offset(offset).
        Find(&tempData).Error; err != nil {
        return nil, err
    }

    combinedData := make([]models.CombinedSensorData, 0, len(tempData))
    for _, temp := range tempData {
        var humid models.SensorHumidity
        res := s.db.Where("timestamp <= ?", temp.Timestamp.Add(5*time.Second).UTC()).
            Order("timestamp DESC").
            First(&humid)
        if res.Error != nil {
            if errors.Is(res.Error, gorm.ErrRecordNotFound) {
                humid.Humidity = 0
            } else {
                return nil, res.Error
            }
        }

        combinedData = append(combinedData, models.CombinedSensorData{
            Timestamp:   temp.Timestamp,
            Temperature: temp.Temperature,
            Humidity:    humid.Humidity,
        })
    }

    totalPages := int(math.Ceil(float64(totalTemp) / float64(pageSize)))

    return &models.SensorDataResponse{
        Data:       combinedData,
        Total:      totalTemp,
        Page:       page,
        PageSize:   pageSize,
        TotalPages: totalPages,
    }, nil
}

func (s *sensorAnalyticsService) GetHourlyData(timeRange string) ([]models.HourlyData, error) {
    start, end := parseTimeRange(timeRange)
    startUTC, endUTC := start.UTC(), end.UTC()

    var hourlyData []models.HourlyData

    current := time.Date(startUTC.Year(), startUTC.Month(), startUTC.Day(), startUTC.Hour(), 0, 0, 0, time.UTC)
    for current.Before(endUTC) {
        hourStart := current
        hourEnd := hourStart.Add(1 * time.Hour)

        var tempData []models.SensorTemperature
        if err := s.db.Where("timestamp BETWEEN ? AND ?", hourStart, hourEnd).
            Find(&tempData).Error; err != nil {
            return nil, err
        }

        var humidData []models.SensorHumidity
        if err := s.db.Where("timestamp BETWEEN ? AND ?", hourStart, hourEnd).
            Find(&humidData).Error; err != nil {
            return nil, err
        }

        if len(tempData) > 0 || len(humidData) > 0 {
            var avgTemp, minTemp, maxTemp float64
            if len(tempData) > 0 {
                var sum float64
                minTemp = tempData[0].Temperature
                maxTemp = tempData[0].Temperature
                for _, t := range tempData {
                    sum += t.Temperature
                    if t.Temperature < minTemp {
                        minTemp = t.Temperature
                    }
                    if t.Temperature > maxTemp {
                        maxTemp = t.Temperature
                    }
                }
                avgTemp = sum / float64(len(tempData))
            }

            var avgHumid, minHumid, maxHumid float64
            if len(humidData) > 0 {
                var sum float64
                minHumid = humidData[0].Humidity
                maxHumid = humidData[0].Humidity
                for _, h := range humidData {
                    sum += h.Humidity
                    if h.Humidity < minHumid {
                        minHumid = h.Humidity
                    }
                    if h.Humidity > maxHumid {
                        maxHumid = h.Humidity
                    }
                }
                avgHumid = sum / float64(len(humidData))
            }

            hourlyData = append(hourlyData, models.HourlyData{
                Hour:        hourStart.Hour(),
                AvgTemp:     math.Round(avgTemp*100) / 100,
                AvgHumidity: math.Round(avgHumid*100) / 100,
                MinTemp:     math.Round(minTemp*100) / 100,
                MaxTemp:     math.Round(maxTemp*100) / 100,
                MinHumidity: math.Round(minHumid*100) / 100,
                MaxHumidity: math.Round(maxHumid*100) / 100,
                Count:       int(math.Max(float64(len(tempData)), float64(len(humidData)))),
            })
        }

        current = current.Add(1 * time.Hour)
    }

    return hourlyData, nil
}