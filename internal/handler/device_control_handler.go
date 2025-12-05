package handler

import (
	"encoding/json"
	"log"
	
	// Pastikan model DeviceControl sudah ada di file models Anda
	"smarthome-backend/database/models"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
)

type DeviceControlHandler struct {
	mqttClient mqtt.Client
}

func NewDeviceControlHandler(client mqtt.Client) *DeviceControlHandler {
	return &DeviceControlHandler{mqttClient: client}
}

// ==================== HELPER FUNCTION (Agar codingan rapi) ====================
func (h *DeviceControlHandler) publishToMQTT(topic string, payload interface{}) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	// Publish dengan QoS 1 agar lebih reliable
	token := h.mqttClient.Publish(topic, 1, false, jsonPayload)
	token.Wait()
	return token.Error()
}

// ==================== ENDPOINTS ====================

// 1. Control - Universal Endpoint (Optional)
func (h *DeviceControlHandler) Control(c *gin.Context) {
	var req models.DeviceControl
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	var topic string
	switch req.Device {
	case "door":
		topic = "iotcihuy/home/door/control"
	case "lamp":
		topic = "iotcihuy/home/lamp/control"
	case "curtain":
		topic = "iotcihuy/home/curtain/control"
	default:
		c.JSON(400, gin.H{"success": false, "error": "Invalid device"})
		return
	}

	payload := map[string]string{"action": req.Action}

	if err := h.publishToMQTT(topic, payload); err != nil {
		log.Printf("❌ MQTT Error: %v", err)
		c.JSON(500, gin.H{"success": false, "error": "Failed to control device"})
		return
	}

	log.Printf("✅ Device control sent: %s -> %s", req.Device, req.Action)
	c.JSON(200, gin.H{"success": true, "message": "Command sent"})
}

// 2. Control Door (Pintu)
func (h *DeviceControlHandler) ControlDoor(c *gin.Context) {
	var req struct {
		Action string `json:"action" binding:"required,oneof=lock unlock"`
		Method string `json:"method"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	if req.Method == "" { req.Method = "remote" }

	payload := map[string]string{
		"action": req.Action,
		"method": req.Method,
	}
	
	// FIX TOPIC
	topic := "iotcihuy/home/door/control"

	if err := h.publishToMQTT(topic, payload); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "Failed to control door"})
		return
	}

	log.Printf("🚪 Door control: %s", req.Action)
	c.JSON(200, gin.H{"success": true, "message": "Door " + req.Action + "ed"})
}

// 3. Control Lamp (Lampu)
func (h *DeviceControlHandler) ControlLamp(c *gin.Context) {
	var req struct {
		Action string `json:"action" binding:"required,oneof=on off"`
		Mode   string `json:"mode"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}
	if req.Mode == "" { req.Mode = "manual" }

	payload := map[string]string{
		"action": req.Action,
		"mode":   req.Mode,
	}
	
	// FIX TOPIC
	topic := "iotcihuy/home/lamp/control"

	if err := h.publishToMQTT(topic, payload); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "Failed to control lamp"})
		return
	}

	log.Printf("💡 Lamp control: %s (%s)", req.Action, req.Mode)
	c.JSON(200, gin.H{"success": true, "message": "Lamp turned " + req.Action})
}

// 4. Control Curtain (Gorden)
func (h *DeviceControlHandler) ControlCurtain(c *gin.Context) {
	var req struct {
		Position int    `json:"position" binding:"required,min=0,max=100"`
		Mode     string `json:"mode"`
		Action   string `json:"action"` // open/close (optional helper)
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}
	if req.Mode == "" { req.Mode = "manual" }
	
	// Helper logic
	if req.Action == "open" { req.Position = 180 }
	if req.Action == "close" { req.Position = 0 }

	payload := map[string]interface{}{
		"position": req.Position,
		"mode":     req.Mode,
		"action":   req.Action,
	}
	
	// FIX TOPIC
	topic := "iotcihuy/home/curtain/control"

	if err := h.publishToMQTT(topic, payload); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "Failed to control curtain"})
		return
	}

	log.Printf("🪟 Curtain control: position %d%% (%s)", req.Position, req.Mode)
	c.JSON(200, gin.H{"success": true, "message": "Curtain position set"})
}

// 5. Control Buzzer (Manual Alert) - INI YANG BARU
func (h *DeviceControlHandler) ControlBuzzer(c *gin.Context) {
	var req struct {
		Action string `json:"action" binding:"required,oneof=on off"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	payload := map[string]string{
		"action": req.Action,
		"source": "manual_api",
	}

	// FIX TOPIC
	topic := "iotcihuy/home/buzzer/control"

	if err := h.publishToMQTT(topic, payload); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "Failed to control buzzer"})
		return
	}

	log.Printf("🚨 Buzzer Manual Control: %s", req.Action)
	c.JSON(200, gin.H{"success": true, "message": "Buzzer turned " + req.Action})
}