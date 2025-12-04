package handler

import (
	"encoding/json"
	"log"
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

// Control - Universal device control endpoint
func (h *DeviceControlHandler) Control(c *gin.Context) {
	var req models.DeviceControl
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Determine MQTT topic based on device
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

	// Create payload
	payload := map[string]string{
		"action": req.Action,
	}
	payloadJSON, _ := json.Marshal(payload)

	// Publish to MQTT
	token := h.mqttClient.Publish(topic, 0, false, payloadJSON)
	token.Wait()

	if token.Error() != nil {
		log.Printf("❌ Failed to publish to MQTT: %v", token.Error())
		c.JSON(500, gin.H{"success": false, "error": "Failed to control device"})
		return
	}

	log.Printf("✅ Device control sent: %s -> %s", req.Device, req.Action)
	c.JSON(200, gin.H{
		"success": true,
		"message": "Device control command sent",
		"device":  req.Device,
		"action":  req.Action,
	})
}

// ControlDoor - Specific endpoint for door control
func (h *DeviceControlHandler) ControlDoor(c *gin.Context) {
	var req struct {
		Action string `json:"action" binding:"required,oneof=lock unlock"`
		Method string `json:"method"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Default method if not provided
	if req.Method == "" {
		req.Method = "remote"
	}

	payload, _ := json.Marshal(map[string]string{
		"action": req.Action,
		"method": req.Method,
	})
	token := h.mqttClient.Publish("iotcihuy/home/door/control", 0, false, payload)
	token.Wait()

	if token.Error() != nil {
		c.JSON(500, gin.H{"success": false, "error": "Failed to control door"})
		return
	}

	log.Printf("🚪 Door control: %s", req.Action)
	c.JSON(200, gin.H{"success": true, "message": "Door " + req.Action + "ed"})
}

// ControlLamp - Specific endpoint for lamp control
func (h *DeviceControlHandler) ControlLamp(c *gin.Context) {
	var req struct {
		Action string `json:"action" binding:"required,oneof=on off"`
		Mode   string `json:"mode" binding:"required,oneof=auto manual"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	payload, _ := json.Marshal(map[string]string{
		"action": req.Action,
		"mode":   req.Mode,
	})
	token := h.mqttClient.Publish("iotcihuy/home/lamp/control", 0, false, payload)
	token.Wait()

	if token.Error() != nil {
		c.JSON(500, gin.H{"success": false, "error": "Failed to control lamp"})
		return
	}

	log.Printf("💡 Lamp control: %s (%s)", req.Action, req.Mode)
	c.JSON(200, gin.H{"success": true, "message": "Lamp turned " + req.Action})
}

// ControlCurtain - Specific endpoint for curtain control
func (h *DeviceControlHandler) ControlCurtain(c *gin.Context) {
	var req struct {
		Position int    `json:"position" binding:"required,min=0,max=100"`
		Mode     string `json:"mode" binding:"required,oneof=auto manual"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"position": req.Position,
		"mode":     req.Mode,
	})
	token := h.mqttClient.Publish("iotcihuy/home/curtain/control", 0, false, payload)
	token.Wait()

	if token.Error() != nil {
		c.JSON(500, gin.H{"success": false, "error": "Failed to control curtain"})
		return
	}

	log.Printf("🪟 Curtain control: position %d%% (%s)", req.Position, req.Mode)
	c.JSON(200, gin.H{"success": true, "message": "Curtain position set"})
}
