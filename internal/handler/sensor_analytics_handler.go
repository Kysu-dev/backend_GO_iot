package handler

import (
	"net/http"
	"smarthome-backend/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SensorAnalyticsHandler struct {
	svc service.SensorAnalyticsService
}

func NewSensorAnalyticsHandler(svc service.SensorAnalyticsService) *SensorAnalyticsHandler {
	return &SensorAnalyticsHandler{svc: svc}
}

// GetStatistics handles GET /api/sensor/stats?range=24h
func (h *SensorAnalyticsHandler) GetStatistics(c *gin.Context) {
	timeRange := c.DefaultQuery("range", "24h")

	stats, err := h.svc.GetStatistics(timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get statistics: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetPaginatedData handles GET /api/sensor/data?range=24h&page=1&page_size=50
func (h *SensorAnalyticsHandler) GetPaginatedData(c *gin.Context) {
	timeRange := c.DefaultQuery("range", "24h")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 50
	}

	data, err := h.svc.GetPaginatedData(timeRange, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get sensor data: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// GetHourlyData handles GET /api/sensor/hourly?range=24h
func (h *SensorAnalyticsHandler) GetHourlyData(c *gin.Context) {
	timeRange := c.DefaultQuery("range", "24h")

	hourlyData, err := h.svc.GetHourlyData(timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get hourly data: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    hourlyData,
	})
}
