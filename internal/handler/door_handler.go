package handler

import (
	"encoding/json"
	"log"
	"smarthome-backend/database/models"
	"smarthome-backend/internal/service"
	"strconv"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
)

type DoorHandler struct {
	svc        service.DoorService
	pinSvc     service.PinService
	mqttClient mqtt.Client
}

func NewDoorHandler(s service.DoorService, p service.PinService, mqttClient mqtt.Client) *DoorHandler {
	return &DoorHandler{
		svc:        s,
		pinSvc:     p,
		mqttClient: mqttClient,
	}
}

func (h *DoorHandler) Create(c *gin.Context) {
	var req models.DoorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	err := h.svc.ProcessDoor(req.Status, req.Method, req.UserID)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "Door status updated"})
}

func (h *DoorHandler) GetLatest(c *gin.Context) {
	data, err := h.svc.GetLatest()
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "No data found"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": data})
}

func (h *DoorHandler) GetAll(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	data, err := h.svc.GetHistory(limit)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "Failed to retrieve data"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": data})
}

// VerifyPin - Validate PIN code from ESP32 keypad and send unlock command via MQTT
func (h *DoorHandler) VerifyPin(c *gin.Context) {
	var req struct {
		Pin string `json:"pin" binding:"required,min=4,max=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "Invalid PIN format"})
		return
	}

	// Get universal PIN from database
	pinData, err := h.pinSvc.GetUniversalPin()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "Failed to retrieve PIN"})
		return
	}

	// Verify PIN
	if req.Pin != pinData.UniversalPin {
		log.Printf("[VERIFY] ❌ Invalid PIN attempt: %s", req.Pin)
		c.JSON(401, gin.H{"success": false, "error": "Invalid PIN", "valid": false})
		return
	}

	log.Printf("[VERIFY] ✅ Valid PIN: Sending unlock command")

	// Send unlock command via MQTT
	topic := "iotcihuy/home/door/control"
	payload := map[string]string{
		"action": "unlock",
		"method": "pin",
	}
	jsonPayload, _ := json.Marshal(payload)

	token := h.mqttClient.Publish(topic, 1, false, jsonPayload)
	token.Wait()

	if token.Error() != nil {
		log.Printf("[ERROR] Failed to publish door unlock: %v", token.Error())
		c.JSON(500, gin.H{"success": false, "error": "Failed to send unlock command"})
		return
	}

	log.Printf("[CONTROL] 🔓 Door → unlock (via PIN)")

	// Save access log to database (async)
	go h.svc.ProcessDoor("unlocked", "pin", nil)

	c.JSON(200, gin.H{
		"success": true,
		"message": "PIN verified successfully, door unlocked",
		"valid":   true,
	})
}
