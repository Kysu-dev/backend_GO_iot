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
	client.Subscribe("iotcihuy/home/gas", 0, h.handleGas)
	client.Subscribe("iotcihuy/home/temperature", 0, h.handleTemperature)
	client.Subscribe("iotcihuy/home/humidity", 0, h.handleHumidity)
	client.Subscribe("iotcihuy/home/light", 0, h.handleLight)

	// Device status subscriptions
	client.Subscribe("iotcihuy/home/door/status", 0, h.handleDoorStatus)
	client.Subscribe("iotcihuy/home/lamp/status", 0, h.handleLampStatus)
	client.Subscribe("iotcihuy/home/curtain/status", 0, h.handleCurtainStatus)

	log.Println("✅ MQTT subscriptions setup complete")
}

// --- HANDLER SENSOR GAS ---
func (h *MQTTHandler) handleGas(c mqtt.Client, m mqtt.Message) {
	var req struct {
		PPM int `json:"ppm"`
	}
	// Cek Error JSON
	if err := json.Unmarshal(m.Payload(), &req); err != nil {
		log.Printf("[MQTT Error] Gas JSON Invalid: %v", err)
		return
	}

	// Simpan DB (Cek Error DB)
	go func() {
		if err := h.gasSvc.ProcessGas(req.PPM); err != nil {
			log.Printf("[DB Error] Gagal simpan Gas: %v", err)
		}
	}()

	// Broadcast
	h.wsHub.BroadcastData(m.Payload())
	log.Printf("[MQTT] Gas: %d", req.PPM)
}

// --- HANDLER TEMPERATURE ---
func (h *MQTTHandler) handleTemperature(c mqtt.Client, m mqtt.Message) {
	var req struct {
		Temperature float64 `json:"temperature"`
	}
	if err := json.Unmarshal(m.Payload(), &req); err != nil {
		log.Printf("[MQTT Error] Temp JSON Invalid: %v", err)
		return
	}

	go func() {
		if err := h.tempSvc.ProcessTemp(req.Temperature); err != nil {
			log.Printf("[DB Error] Gagal simpan Temp: %v", err)
		}
	}()

	h.wsHub.BroadcastData(m.Payload())
	log.Printf("[MQTT] Temperature: %.2f", req.Temperature)
}

// --- HANDLER HUMIDITY ---
func (h *MQTTHandler) handleHumidity(c mqtt.Client, m mqtt.Message) {
	var req struct {
		Humidity float64 `json:"humidity"`
	}
	if err := json.Unmarshal(m.Payload(), &req); err != nil {
		log.Printf("[MQTT Error] Humid JSON Invalid: %v", err)
		return
	}

	go func() {
		if err := h.humidSvc.ProcessHumid(req.Humidity); err != nil {
			log.Printf("[DB Error] Gagal simpan Humid: %v", err)
		}
	}()

	h.wsHub.BroadcastData(m.Payload())
	log.Printf("[MQTT] Humidity: %.2f", req.Humidity)
}

// --- HANDLER LIGHT ---
func (h *MQTTHandler) handleLight(c mqtt.Client, m mqtt.Message) {
	var req struct {
		Lux int `json:"lux"`
	}
	if err := json.Unmarshal(m.Payload(), &req); err != nil {
		log.Printf("[MQTT Error] Light JSON Invalid: %v", err)
		return
	}

	go func() {
		if err := h.lightSvc.ProcessLight(req.Lux); err != nil {
			log.Printf("[DB Error] Gagal simpan Light: %v", err)
		}
	}()

	h.wsHub.BroadcastData(m.Payload())
	log.Printf("[MQTT] Light: %d", req.Lux)
}

// --- HANDLER DOOR ---
func (h *MQTTHandler) handleDoorStatus(c mqtt.Client, m mqtt.Message) {
	var req struct {
		Status string `json:"status"`
		Method string `json:"method"`
	}
	if err := json.Unmarshal(m.Payload(), &req); err != nil {
		log.Printf("[MQTT Error] Door JSON Invalid: %v", err)
		return
	}

	go func() {
		if err := h.doorSvc.ProcessDoor(req.Status, req.Method); err != nil {
			log.Printf("[DB Error] Gagal simpan Door: %v", err)
		}
	}()

	h.wsHub.BroadcastData(m.Payload())
	log.Printf("[MQTT] Door: %s via %s", req.Status, req.Method)
}

// --- HANDLER LAMP ---
func (h *MQTTHandler) handleLampStatus(c mqtt.Client, m mqtt.Message) {
	var req struct {
		Status string `json:"status"`
		Mode   string `json:"mode"`
	}
	if err := json.Unmarshal(m.Payload(), &req); err != nil {
		log.Printf("[MQTT Error] Lamp JSON Invalid: %v", err)
		return
	}

	go func() {
		if err := h.lampSvc.ProcessLamp(req.Status, req.Mode); err != nil {
			log.Printf("[DB Error] Gagal simpan Lamp: %v", err)
		}
	}()

	h.wsHub.BroadcastData(m.Payload())
	log.Printf("[MQTT] Lamp: %s (%s)", req.Status, req.Mode)
}

// --- HANDLER CURTAIN ---
func (h *MQTTHandler) handleCurtainStatus(c mqtt.Client, m mqtt.Message) {
	var req struct {
		Position int    `json:"position"`
		Mode     string `json:"mode"`
	}
	if err := json.Unmarshal(m.Payload(), &req); err != nil {
		log.Printf("[MQTT Error] Curtain JSON Invalid: %v", err)
		return
	}

	go func() {
		if err := h.curtainSvc.ProcessCurtain(req.Position, req.Mode); err != nil {
			log.Printf("[DB Error] Gagal simpan Curtain: %v", err)
		}
	}()

	h.wsHub.BroadcastData(m.Payload())
	log.Printf("[MQTT] Curtain: %d%% (%s)", req.Position, req.Mode)
}
