package mqtt

import (
	"encoding/json"
	"log"
	"smarthome-backend/internal/service"
	"smarthome-backend/internal/websocket"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTHandler struct {
	client     mqtt.Client // <--- 1. TAMBAHAN: Client disimpan disini agar bisa Publish
	gasSvc     service.GasService
	tempSvc    service.TempService
	humidSvc   service.HumidService
	lightSvc   service.LightService
	doorSvc    service.DoorService
	lampSvc    service.LampService
	curtainSvc service.CurtainService
	wsHub      *websocket.Hub
}

// Update Constructor: Menerima 'client' sebagai parameter pertama
func NewMQTTHandler(
	client mqtt.Client, // <--- 2. TAMBAHAN: Parameter Client
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
		client:     client, // <--- Simpan ke struct
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
	// Topic Subscribe (Mendengar Data dari ESP32)
	topics := map[string]mqtt.MessageHandler{
		"iotcihuy/home/temperature":   h.handleTemperature,
		"iotcihuy/home/humidity":      h.handleHumidity,
		"iotcihuy/home/gas":           h.handleGas,
		"iotcihuy/home/light":         h.handleLight,
		"iotcihuy/home/lamp/status":   h.handleLampStatus,
		"iotcihuy/home/door/status":   h.handleDoorStatus,
		"iotcihuy/home/curtain/status": h.handleCurtainStatus,
	}

	for topic, handler := range topics {
		if token := client.Subscribe(topic, 0, handler); token.Wait() && token.Error() != nil {
			log.Printf("❌ Failed to subscribe to %s: %v", topic, token.Error())
		}
	}
	log.Println("✅ MQTT subscriptions setup complete")
}

// ==================== PUBLISH FUNCTION (UNTUK TOMBOL PINTU) ====================

// Fungsi ini dipanggil dari HTTP Handler (Button di klik)
func (h *MQTTHandler) PublishDoorControl(action string) error {
	// Topik khusus untuk memerintah ESP32
	topic := "iotcihuy/home/door/control"

	// Payload: {"action": "lock"} atau {"action": "unlock"}
	payload := map[string]string{
		"action": action,
		"method": "remote",
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// PUBLISH ke MQTT
	token := h.client.Publish(topic, 1, false, jsonPayload)
	token.Wait()

	if token.Error() != nil {
		log.Printf("❌ Gagal Publish Door Control: %v", token.Error())
		return token.Error()
	}

	log.Printf("📤 [MQTT Publish] Kirim Perintah Pintu: %s ke topik %s", action, topic)
	return nil
}

// ==================== SENSOR HANDLERS (SUBSCRIBE) ====================

func (h *MQTTHandler) handleTemperature(client mqtt.Client, msg mqtt.Message) {
	var data struct {
		Temperature float64 `json:"temperature"`
		Unit        string  `json:"unit"`
	}

	if err := json.Unmarshal(msg.Payload(), &data); err != nil {
		log.Printf("❌ Error parsing temperature: %v", err)
		return
	}

	go func() {
		if err := h.tempSvc.ProcessTemp(data.Temperature); err != nil {
			log.Printf("❌ Error saving temperature: %v", err)
		}
	}()

	h.wsHub.BroadcastData(msg.Payload())
	log.Printf("[MQTT] Temperature: %.2f°C", data.Temperature)
}

func (h *MQTTHandler) handleHumidity(client mqtt.Client, msg mqtt.Message) {
	var data struct {
		Humidity float64 `json:"humidity"`
		Unit     string  `json:"unit"`
	}

	if err := json.Unmarshal(msg.Payload(), &data); err != nil {
		log.Printf("❌ Error parsing humidity: %v", err)
		return
	}

	go func() {
		if err := h.humidSvc.ProcessHumid(data.Humidity); err != nil {
			log.Printf("❌ Error saving humidity: %v", err)
		}
	}()

	h.wsHub.BroadcastData(msg.Payload())
	log.Printf("[MQTT] Humidity: %.2f%%", data.Humidity)
}

func (h *MQTTHandler) handleGas(client mqtt.Client, msg mqtt.Message) {
	var data struct {
		PPM  int    `json:"ppm"`
		Unit string `json:"unit"`
	}
	if err := json.Unmarshal(msg.Payload(), &data); err != nil {
		log.Printf("❌ Error parsing gas: %v", err)
		return
	}

	go func() {
		if err := h.gasSvc.ProcessGas(data.PPM); err != nil {
			log.Printf("❌ Error saving gas: %v", err)
		}
	}()

	h.wsHub.BroadcastData(msg.Payload())
	log.Printf("[MQTT] Gas: %d PPM", data.PPM)
}

func (h *MQTTHandler) handleLight(client mqtt.Client, msg mqtt.Message) {
	var data struct {
		Lux  int    `json:"lux"`
		Unit string `json:"unit"`
	}
	if err := json.Unmarshal(msg.Payload(), &data); err != nil {
		log.Printf("❌ Error parsing light: %v", err)
		return
	}

	go func() {
		if err := h.lightSvc.ProcessLight(data.Lux); err != nil {
			log.Printf("❌ Error saving light: %v", err)
		}
	}()

	h.wsHub.BroadcastData(msg.Payload())
	log.Printf("[MQTT] Light: %d Lux", data.Lux)
}

// ==================== DEVICE HANDLERS (STATUS) ====================

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

	// Broadcast format WebSocket khusus Device
	wsData := map[string]interface{}{
		"type":   "device_update",
		"device": "door",
		"locked": req.Status == "locked",
		"method": req.Method,
	}
	jsonData, _ := json.Marshal(wsData)
	h.wsHub.BroadcastData(jsonData)

	log.Printf("[MQTT] Door Status: %s via %s", req.Status, req.Method)
}

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

	wsData := map[string]interface{}{
		"type":   "device_update",
		"device": "lamp",
		"state":  req.Status == "on",
		"mode":   req.Mode,
	}
	jsonData, _ := json.Marshal(wsData)
	h.wsHub.BroadcastData(jsonData)
	log.Printf("[MQTT] Lamp: %s (%s)", req.Status, req.Mode)
}

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

	wsData := map[string]interface{}{
		"type":     "device_update",
		"device":   "curtain",
		"position": req.Position,
		"mode":     req.Mode,
	}
	jsonData, _ := json.Marshal(wsData)
	h.wsHub.BroadcastData(jsonData)
	log.Printf("[MQTT] Curtain: %d%% (%s)", req.Position, req.Mode)
}