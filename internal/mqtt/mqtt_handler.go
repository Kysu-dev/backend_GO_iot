package mqtt

import (
	"encoding/json"
	"log"
	"smarthome-backend/internal/service"
	"smarthome-backend/internal/websocket" // Import WebSocket
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTHandler struct {
	gasSvc service.GasService
	wsHub  *websocket.Hub
}

func NewMQTTHandler(g service.GasService, hub *websocket.Hub) *MQTTHandler {
	return &MQTTHandler{gasSvc: g, wsHub: hub}
}

func (h *MQTTHandler) SetupRoutes(client mqtt.Client) {
	client.Subscribe("home/gas", 0, h.handleGas)
}

func (h *MQTTHandler) handleGas(c mqtt.Client, m mqtt.Message) {
	// 1. Parsing Data dari Sensor
	var req struct { PPM int `json:"ppm"` }
	json.Unmarshal(m.Payload(), &req)

	// 2. Simpan ke Database
	go h.gasSvc.ProcessGas(req.PPM)
	
	// 3. KIRIM KE WEBSOCKET (REALTIME)
	h.wsHub.BroadcastData(m.Payload())

	log.Printf("[MQTT] Gas: %d (Saved & Broadcasted)", req.PPM)
}