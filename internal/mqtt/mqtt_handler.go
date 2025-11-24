package mqtt

import (
	"encoding/json"
	"log"
	"smarthome-backend/internal/service"
	"smarthome-backend/internal/websocket" // Import WebSocket

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTHandler struct {
	gasSvc   service.GasService
	tempSvc  service.TempService
	humidSvc service.HumidService
	lightSvc service.LightService
	wsHub    *websocket.Hub
}

func NewMQTTHandler(g service.GasService, t service.TempService, h service.HumidService, l service.LightService, hub *websocket.Hub) *MQTTHandler {
	return &MQTTHandler{
		gasSvc:   g,
		tempSvc:  t,
		humidSvc: h,
		lightSvc: l,
		wsHub:    hub,
	}
}

func (h *MQTTHandler) SetupRoutes(client mqtt.Client) {
	client.Subscribe("home/gas", 0, h.handleGas)
	client.Subscribe("home/temperature", 0, h.handleTemperature)
	client.Subscribe("home/humidity", 0, h.handleHumidity)
	client.Subscribe("home/light", 0, h.handleLight)
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
