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
	// Sensor subscriptions with error checking
	log.Println("🔄 Setting up MQTT subscriptions...")

	if token := client.Subscribe("home/gas", 0, h.handleGas); token.Wait() && token.Error() != nil {
		log.Printf("❌ Failed to subscribe to home/gas: %v", token.Error())
	} else {
		log.Println("✅ Subscribed to: home/gas")
	}

	if token := client.Subscribe("home/temperature", 0, h.handleTemperature); token.Wait() && token.Error() != nil {
		log.Printf("❌ Failed to subscribe to home/temperature: %v", token.Error())
	} else {
		log.Println("✅ Subscribed to: home/temperature")
	}

	if token := client.Subscribe("home/humidity", 0, h.handleHumidity); token.Wait() && token.Error() != nil {
		log.Printf("❌ Failed to subscribe to home/humidity: %v", token.Error())
	} else {
		log.Println("✅ Subscribed to: home/humidity")
	}

	if token := client.Subscribe("home/light", 0, h.handleLight); token.Wait() && token.Error() != nil {
		log.Printf("❌ Failed to subscribe to home/light: %v", token.Error())
	} else {
		log.Println("✅ Subscribed to: home/light")
	}

	// Device status subscriptions
	if token := client.Subscribe("home/door/status", 0, h.handleDoorStatus); token.Wait() && token.Error() != nil {
		log.Printf("❌ Failed to subscribe to home/door/status: %v", token.Error())
	} else {
		log.Println("✅ Subscribed to: home/door/status")
	}

	if token := client.Subscribe("home/lamp/status", 0, h.handleLampStatus); token.Wait() && token.Error() != nil {
		log.Printf("❌ Failed to subscribe to home/lamp/status: %v", token.Error())
	} else {
		log.Println("✅ Subscribed to: home/lamp/status")
	}

	if token := client.Subscribe("home/curtain/status", 0, h.handleCurtainStatus); token.Wait() && token.Error() != nil {
		log.Printf("❌ Failed to subscribe to home/curtain/status: %v", token.Error())
	} else {
		log.Println("✅ Subscribed to: home/curtain/status")
	}

	log.Println("✅ MQTT subscriptions setup complete")
}

func (h *MQTTHandler) handleGas(c mqtt.Client, m mqtt.Message) {
	log.Printf("📩 Received MQTT on topic: %s, payload: %s", m.Topic(), string(m.Payload()))

	// 1. Parsing Data dari Sensor
	var req struct {
		PPM int `json:"ppm"`
	}
	if err := json.Unmarshal(m.Payload(), &req); err != nil {
		log.Printf("❌ Failed to parse gas data: %v", err)
		return
	}

	// 2. Simpan ke Database
	go h.gasSvc.ProcessGas(req.PPM)

	// 3. KIRIM KE WEBSOCKET (REALTIME)
	h.wsHub.BroadcastData(m.Payload())

	log.Printf("[MQTT] Gas: %d (Saved & Broadcasted)", req.PPM)
}

func (h *MQTTHandler) handleTemperature(c mqtt.Client, m mqtt.Message) {
	log.Printf("📩 Received MQTT on topic: %s, payload: %s", m.Topic(), string(m.Payload()))

	// 1. Parsing Data dari Sensor
	var req struct {
		Temperature float64 `json:"temperature"`
	}
	if err := json.Unmarshal(m.Payload(), &req); err != nil {
		log.Printf("❌ Failed to parse temperature data: %v", err)
		return
	}

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

	log.Printf("[MQTT] Door: %s via %s (Saved & Broadcasted)", req.Status, req.Method)
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

	log.Printf("[MQTT] Lamp: %s (%s mode) (Saved & Broadcasted)", req.Status, req.Mode)
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

	log.Printf("[MQTT] Curtain: position %d%% (%s mode) (Saved & Broadcasted)", req.Position, req.Mode)
}
