package handler

import (
	"net/http"
	"smarthome-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	tempSvc    service.TempService
	humidSvc   service.HumidService
	gasSvc     service.GasService
	lightSvc   service.LightService
	lampSvc    service.LampService
	doorSvc    service.DoorService
	curtainSvc service.CurtainService
}

func NewDashboardHandler(
	tempSvc service.TempService,
	humidSvc service.HumidService,
	gasSvc service.GasService,
	lightSvc service.LightService,
	lampSvc service.LampService,
	doorSvc service.DoorService,
	curtainSvc service.CurtainService,
) *DashboardHandler {
	return &DashboardHandler{
		tempSvc:    tempSvc,
		humidSvc:   humidSvc,
		gasSvc:     gasSvc,
		lightSvc:   lightSvc,
		lampSvc:    lampSvc,
		doorSvc:    doorSvc,
		curtainSvc: curtainSvc,
	}
}

// GetInitialData returns aggregated dashboard data (sensors + devices)
func (h *DashboardHandler) GetInitialData(c *gin.Context) {
	// Fetch latest sensor data (limit 1 for latest only)
	temps, _ := h.tempSvc.GetHistory(1)
	humids, _ := h.humidSvc.GetHistory(1)
	gases, _ := h.gasSvc.GetHistory(1)
	lights, _ := h.lightSvc.GetHistory(1)

	// Fetch latest device status
	lamp, _ := h.lampSvc.GetLatest()
	door, _ := h.doorSvc.GetLatest()
	curtain, _ := h.curtainSvc.GetLatest()

	// Build sensor data
	sensorData := map[string]interface{}{
		"temperature": 0.0,
		"humidity":    0.0,
		"gas":         0,
		"gas_ppm":     0,
		"light":       0,
	}

	if len(temps) > 0 {
		sensorData["temperature"] = temps[0].Temperature
	}
	if len(humids) > 0 {
		sensorData["humidity"] = humids[0].Humidity
	}
	if len(gases) > 0 {
		sensorData["gas"] = gases[0].PPMValue
		sensorData["gas_ppm"] = gases[0].PPMValue
	}
	if len(lights) > 0 {
		sensorData["light"] = lights[0].Lux
	}

	// Build device data
	deviceData := map[string]interface{}{
		"lamp":    "off",
		"door":    "locked",
		"curtain": "closed",
	}

	if lamp != nil {
		deviceData["lamp"] = lamp.Status
	}
	if door != nil {
		deviceData["door"] = door.Status
	}
	if curtain != nil {
		deviceData["curtain"] = curtain.Status
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"sensors": sensorData,
			"devices": deviceData,
		},
	})
}
