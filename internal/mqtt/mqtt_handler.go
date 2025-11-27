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
	gasSvc service.GasService,
	tempSvc service.TempService,
	humidSvc service.HumidService,
	lightSvc service.LightService,
	doorSvc service.DoorService,
	lampSvc service.LampService,
	curtainSvc service.CurtainService,
	wsHub *websocket.Hub,
) *MQTTHandler {
	return &MQTTHandler{
		gasSvc:     gasSvc,
		tempSvc:    tempSvc,
		humidSvc:   humidSvc,
		lightSvc:   lightSvc,
		doorSvc:    doorSvc,
		lampSvc:    lampSvc,
		curtainSvc: curtainSvc,
		wsHub:      wsHub,
	}
}

func (h *MQTTHandler) SetupRoutes(client mqtt.Client) {
	topics := map[string]mqtt.MessageHandler{
		"iotcihuy/home/temperature": h.handleTemperature,
		"iotcihuy/home/humidity":    h.handleHumidity,
		"iotcihuy/home/gas":         h.handleGas,
		"iotcihuy/home/light":       h.handleLight,
		"iotcihuy/home/lamp/status": h.handleLampStatus,
		"iotcihuy/home/door/status": h.handleDoorStatus,
	}

	for topic, handler := range topics {
		if token := client.Subscribe(topic, 0, handler); token.Wait() && token.Error() != nil {
			log.Printf("❌ Failed to subscribe to %s: %v", topic, token.Error())
		}
	}
	log.Println("✅ MQTT subscriptions setup complete")
}

// ==================== SENSOR HANDLERS ====================

func (h *MQTTHandler) handleTemperature(client mqtt.Client, msg mqtt.Message) {
	var data struct {
		Temperature float64 `json:"temperature"`
		Unit        string  `json:"unit"`
	}

	if err := json.Unmarshal(msg.Payload(), &data); err != nil {
		log.Printf("❌ Error parsing temperature: %v", err)
		return
	}

	// ✅ Save to database using ProcessTemp
	go func() {
		if err := h.tempSvc.ProcessTemp(data.Temperature); err != nil {
			log.Printf("❌ Error saving temperature: %v", err)
		}
	}()

	log.Printf("[MQTT] Temperature: %.2f°C", data.Temperature)

	// ⭐ Broadcast to WebSocket
	wsData, _ := json.Marshal(map[string]interface{}{
		"type":   "sensor_update",
		"sensor": "temperature",
		"value":  data.Temperature,
		"unit":   data.Unit,
	})
	h.wsHub.BroadcastData(wsData)
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

	// ✅ Save to database using ProcessHumid
	go func() {
		if err := h.humidSvc.ProcessHumid(data.Humidity); err != nil {
			log.Printf("❌ Error saving humidity: %v", err)
		}
	}()

	log.Printf("[MQTT] Humidity: %.2f%%", data.Humidity)

	// ⭐ Broadcast to WebSocket
	wsData, _ := json.Marshal(map[string]interface{}{
		"type":   "sensor_update",
		"sensor": "humidity",
		"value":  data.Humidity,
		"unit":   data.Unit,
	})
	h.wsHub.BroadcastData(wsData)
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

	// ✅ Save to database using ProcessGas
	go func() {
		if err := h.gasSvc.ProcessGas(data.PPM); err != nil {
			log.Printf("❌ Error saving gas: %v", err)
		}
	}()

	log.Printf("[MQTT] Gas: %d PPM", data.PPM)

	// ⭐ Broadcast to WebSocket
	wsData, _ := json.Marshal(map[string]interface{}{
		"type":   "sensor_update",
		"sensor": "gas",
		"value":  data.PPM,
		"unit":   data.Unit,
	})
	h.wsHub.BroadcastData(wsData)
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

	// ✅ Save to database using ProcessLight
	go func() {
		if err := h.lightSvc.ProcessLight(data.Lux); err != nil {
			log.Printf("❌ Error saving light: %v", err)
		}
	}()

	log.Printf("[MQTT] Light: %d Lux", data.Lux)

	// ⭐ Broadcast to WebSocket
	wsData, _ := json.Marshal(map[string]interface{}{
		"type":   "sensor_update",
		"sensor": "light",
		"value":  data.Lux,
		"unit":   data.Unit,
	})
	h.wsHub.BroadcastData(wsData)
}

// ==================== DEVICE HANDLERS ====================

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

	// Broadcast dengan format WebSocket
	wsData := map[string]interface{}{
		"type":   "device_update",
		"device": "door",
		"locked": req.Status == "locked",
		"method": req.Method,
	}
	jsonData, _ := json.Marshal(wsData)
	h.wsHub.BroadcastData(jsonData)
	log.Printf("[MQTT] Door: %s via %s", req.Status, req.Method)
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

	// Broadcast dengan format WebSocket
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

	// Broadcast dengan format WebSocket
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
