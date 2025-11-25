package mqtt

import (
	"encoding/json"
	"log"
	"smarthome-backend/internal/service"
	"smarthome-backend/internal/websocket"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTHandler struct {
	gasSvc     service.GasService
	tempSvc    service.TempService
	humidSvc   service.HumidService
	lightSvc   service.LightService
	doorSvc    service.DoorService
	lampSvc    service.LampService
	curtainSvc service.CurtainService
	wsHub      *websocket.Hub
}

func NewMQTTHandler(
	g service.GasService,
	t service.TempService,
	h service.HumidService,
	l service.LightService,
	d service.DoorService,
	lamp service.LampService,
	curtain service.CurtainService,
	hub *websocket.Hub,
) *MQTTHandler {
	return &MQTTHandler{
		gasSvc:     g,
		tempSvc:    t,
		humidSvc:   h,
		lightSvc:   l,
		doorSvc:    d,
		lampSvc:    lamp,
		curtainSvc: curtain,
		wsHub:      hub,
	}
}

func (h *MQTTHandler) SetupRoutes(client mqtt.Client) {
	// Sensor subscriptions
	client.Subscribe("home/gas", 0, h.handleGas)
	client.Subscribe("home/temperature", 0, h.handleTemperature)
	client.Subscribe("home/humidity", 0, h.handleHumidity)
	client.Subscribe("home/light", 0, h.handleLight)

	// Device status subscriptions
	client.Subscribe("home/door/status", 0, h.handleDoorStatus)
	client.Subscribe("home/lamp/status", 0, h.handleLampStatus)
	client.Subscribe("home/curtain/status", 0, h.handleCurtainStatus)

	log.Println("✅ MQTT subscriptions setup complete")
}

func (h *MQTTHandler) handleGas(c mqtt.Client, m mqtt.Message) {
	// 1. Parsing Data dari Sensor
	var req struct {
		PPM int `json:"ppm"`
	}
	json.Unmarshal(m.Payload(), &req)

	// 2. Simpan ke Database
	go h.gasSvc.ProcessGas(req.PPM)

	// 3. KIRIM KE WEBSOCKET (REALTIME)
	h.wsHub.BroadcastData(m.Payload())

	log.Printf("[MQTT] Gas: %d (Saved & Broadcasted)", req.PPM)
}

func (h *MQTTHandler) handleTemperature(c mqtt.Client, m mqtt.Message) {
	// 1. Parsing Data dari Sensor
	var req struct {
		Temperature float64 `json:"temperature"`
	}
	json.Unmarshal(m.Payload(), &req)

	// 2. Simpan ke Database
	go h.tempSvc.ProcessTemp(req.Temperature)

	// 3. KIRIM KE WEBSOCKET (REALTIME)
	h.wsHub.BroadcastData(m.Payload())

	log.Printf("[MQTT] Temperature: %.2f (Saved & Broadcasted)", req.Temperature)
}

func (h *MQTTHandler) handleHumidity(c mqtt.Client, m mqtt.Message) {
	// 1. Parsing Data dari Sensor
	var req struct {
		Humidity float64 `json:"humidity"`
	}
	json.Unmarshal(m.Payload(), &req)

	// 2. Simpan ke Database
	go h.humidSvc.ProcessHumid(req.Humidity)

	// 3. KIRIM KE WEBSOCKET (REALTIME)
	h.wsHub.BroadcastData(m.Payload())

	log.Printf("[MQTT] Humidity: %.2f (Saved & Broadcasted)", req.Humidity)
}

func (h *MQTTHandler) handleLight(c mqtt.Client, m mqtt.Message) {
	// 1. Parsing Data dari Sensor
	var req struct {
		Lux int `json:"lux"`
	}
	json.Unmarshal(m.Payload(), &req)

	// 2. Simpan ke Database
	go h.lightSvc.ProcessLight(req.Lux)

	// 3. KIRIM KE WEBSOCKET (REALTIME)
	h.wsHub.BroadcastData(m.Payload())

	log.Printf("[MQTT] Light: %d lux (Saved & Broadcasted)", req.Lux)
}

// ==================== DEVICE STATUS HANDLERS ====================

func (h *MQTTHandler) handleDoorStatus(c mqtt.Client, m mqtt.Message) {
	var req struct {
		Status string `json:"status"`
		Method string `json:"method"`
	}
	json.Unmarshal(m.Payload(), &req)

	// Save to database
	go h.doorSvc.ProcessDoor(req.Status, req.Method)

	// Broadcast to WebSocket
	h.wsHub.BroadcastData(m.Payload())

	log.Printf("[MQTT] Door: %s (method: %s)", req.Status, req.Method)
}

func (h *MQTTHandler) handleLampStatus(c mqtt.Client, m mqtt.Message) {
	var req struct {
		Status string `json:"status"`
		Mode   string `json:"mode"`
	}
	json.Unmarshal(m.Payload(), &req)

	// Save to database
	go h.lampSvc.ProcessLamp(req.Status, req.Mode)

	// Broadcast to WebSocket
	h.wsHub.BroadcastData(m.Payload())

	log.Printf("[MQTT] Lamp: %s (mode: %s)", req.Status, req.Mode)
}

func (h *MQTTHandler) handleCurtainStatus(c mqtt.Client, m mqtt.Message) {
	var req struct {
		Position int    `json:"position"`
		Mode     string `json:"mode"`
	}
	json.Unmarshal(m.Payload(), &req)

	// Save to database
	go h.curtainSvc.ProcessCurtain(req.Position, req.Mode)

	// Broadcast to WebSocket
	h.wsHub.BroadcastData(m.Payload())

	log.Printf("[MQTT] Curtain: position %d%% (mode: %s)", req.Position, req.Mode)
}
