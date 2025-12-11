package handler

import (
	"encoding/json"
	"log"

	"smarthome-backend/internal/service"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
)

type DeviceControlHandler struct {
	mqttClient mqtt.Client
	lampSvc    service.LampService
	doorSvc    service.DoorService
	curtainSvc service.CurtainService
}

// Constructor sesuai dengan main.go (4 Parameter)
func NewDeviceControlHandler(
	client mqtt.Client,
	lampSvc service.LampService,
	doorSvc service.DoorService,
	curtainSvc service.CurtainService,
) *DeviceControlHandler {
	return &DeviceControlHandler{
		mqttClient: client,
		lampSvc:    lampSvc,
		doorSvc:    doorSvc,
		curtainSvc: curtainSvc,
	}
}

// Helper untuk Publish ke MQTT
func (h *DeviceControlHandler) publishToMQTT(topic string, payload interface{}) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	token := h.mqttClient.Publish(topic, 1, false, jsonPayload)
	token.Wait()
	return token.Error()
}

// --- UNIVERSAL CONTROL (Opsional, jika ingin satu endpoint untuk semua) ---
func (h *DeviceControlHandler) Control(c *gin.Context) {
	var req struct {
		Device string                 `json:"device" binding:"required"`
		Action map[string]interface{} `json:"action" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Contoh implementasi sederhana publish raw
	topic := "iotcihuy/home/" + req.Device + "/control"
	if err := h.publishToMQTT(topic, req.Action); err != nil {
		c.JSON(500, gin.H{"error": "Failed to publish MQTT"})
		return
	}

	c.JSON(200, gin.H{"message": "Command sent", "device": req.Device})
}

// 1. CONTROL DOOR (Lock/Unlock)
func (h *DeviceControlHandler) ControlDoor(c *gin.Context) {
	var req struct {
		Action string `json:"action" binding:"required,oneof=lock unlock"`
		Method string `json:"method"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if req.Method == "" {
		req.Method = "remote"
	}

	// A. Kirim ke MQTT (Agar Alat Bergerak)
	payload := map[string]string{"action": req.Action, "method": req.Method}
	if err := h.publishToMQTT("iotcihuy/home/door/control", payload); err != nil {
		log.Printf("❌ MQTT Error: %v", err)
		c.JSON(500, gin.H{"error": "Failed MQTT"})
		return
	}

	// B. SIMPAN KE DB LANGSUNG (Optimistic Update)
	status := "locked"
	if req.Action == "unlock" {
		status = "unlocked"
	}

	h.doorSvc.ProcessDoor(status, req.Method, nil)

	c.JSON(200, gin.H{"success": true, "message": "Door Command Sent & Saved"})
}

// 2. CONTROL LAMP (On/Off + Auto/Manual)
func (h *DeviceControlHandler) ControlLamp(c *gin.Context) {
	var req struct {
		Action string `json:"action" binding:"required,oneof=on off"`
		Mode   string `json:"mode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if req.Mode == "" {
		req.Mode = "manual"
	}

	// A. Kirim ke MQTT
	payload := map[string]string{"action": req.Action, "mode": req.Mode}
	if err := h.publishToMQTT("iotcihuy/home/lamp/control", payload); err != nil {
		log.Printf("❌ MQTT Error: %v", err)
		c.JSON(500, gin.H{"error": "Failed MQTT"})
		return
	}

	// B. SIMPAN KE DB LANGSUNG
	h.lampSvc.ProcessLamp(req.Action, req.Mode)

	c.JSON(200, gin.H{"success": true, "message": "Lamp Command Sent & Saved"})
}

// 3. CONTROL CURTAIN (Open/Close + Auto/Manual)
func (h *DeviceControlHandler) ControlCurtain(c *gin.Context) {
	var req struct {
		Action string `json:"action" binding:"required,oneof=open close"`
		Mode   string `json:"mode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if req.Mode == "" {
		req.Mode = "manual"
	}

	// A. Kirim ke MQTT
	payload := map[string]string{"action": req.Action, "mode": req.Mode}
	if err := h.publishToMQTT("iotcihuy/home/curtain/control", payload); err != nil {
		log.Printf("❌ MQTT Error: %v", err)
		c.JSON(500, gin.H{"error": "Failed MQTT"})
		return
	}

	// B. SIMPAN KE DB LANGSUNG
	status := "closed"
	if req.Action == "open" {
		status = "open"
	}
	h.curtainSvc.ProcessCurtain(status, req.Mode)

	c.JSON(200, gin.H{"success": true, "message": "Curtain Command Sent & Saved"})
}

// 4. CONTROL BUZZER (Manual)
func (h *DeviceControlHandler) ControlBuzzer(c *gin.Context) {
	var req struct {
		Action string `json:"action" binding:"required,oneof=on off"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	payload := map[string]string{"action": req.Action}
	if err := h.publishToMQTT("iotcihuy/home/buzzer/control", payload); err != nil {
		c.JSON(500, gin.H{"error": "Failed MQTT"})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "Buzzer Command Sent"})
}
