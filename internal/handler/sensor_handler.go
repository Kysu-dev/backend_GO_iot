package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"smarthome-backend/internal/service"
)

type SensorRequest struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Gas         int     `json:"gas"`
	Light       int     `json:"light"`
}


func LogEnvironmentData(c *gin.Context) {
	var req SensorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := service.LogSensorData(req.Temperature, req.Humidity, req.Gas, req.Light)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data sensor berhasil disimpan"})
}

// GET /api/
func GetLatestData(c *gin.Context) {
	data, err := service.GetLatestData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data"})
		return
	}
	c.JSON(http.StatusOK, data)
}
