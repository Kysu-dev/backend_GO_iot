package mqtt

import (
	"encoding/json"
	"log"
	"smarthome-backend/internal/service"
	"smarthome-backend/internal/websocket"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTHandler struct {
	client     mqtt.Client
	gasSvc     service.GasService
	tempSvc    service.TempService
	humidSvc   service.HumidService
	lightSvc   service.LightService
	doorSvc    service.DoorService
	lampSvc    service.LampService
	curtainSvc service.CurtainService
	pinSvc     service.PinService
	wsHub      *websocket.Hub

	// Batch sensor persistence
	batchInterval time.Duration
	sensorCache   sensorCache

	// Buzzer state tracking
	lastBuzzerState string
	buzzerMutex     sync.Mutex

	// Lamp state tracking
	lastLampState      string
	lastLampMode       string
	lampMutex          sync.RWMutex
	lastLampChangeTime time.Time

	// Curtain state tracking
	lastCurtainState      string
	lastCurtainMode       string
	curtainMutex          sync.RWMutex
	lastCurtainChangeTime time.Time

	// Gas moving average filter
	gasReadings     []int
	gasReadingMutex sync.Mutex
	maxGasReadings  int
}

func NewMQTTHandler(
	client mqtt.Client,
	g service.GasService,
	t service.TempService,
	h service.HumidService,
	l service.LightService,
	d service.DoorService,
	lamp service.LampService,
	curtain service.CurtainService,
	pin service.PinService,
	hub *websocket.Hub,
) *MQTTHandler {
	handler := &MQTTHandler{
		client:             client,
		gasSvc:             g,
		tempSvc:            t,
		humidSvc:           h,
		lightSvc:           l,
		doorSvc:            d,
		lampSvc:            lamp,
		curtainSvc:         curtain,
		pinSvc:             pin,
		wsHub:              hub,
		batchInterval:      defaultSensorBatchInterval,
		lastBuzzerState:    "off",
		lastLampState:      "off",
		lastLampMode:       "auto",
		lastLampChangeTime: time.Now(),
		lastCurtainState:   "closed",
		lastCurtainMode:    "auto",
		gasReadings:        make([]int, 0, 5), // Buffer 5 readings
		maxGasReadings:     5,
	}

	handler.startSensorBatcher()

	return handler
}

func (h *MQTTHandler) startSensorBatcher() {
	ticker := time.NewTicker(h.batchInterval)
	go func() {
		for range ticker.C {
			h.flushSensorCache()
		}
	}()
}

func (h *MQTTHandler) SetupRoutes(client mqtt.Client) {
	h.client = client

	topics := map[string]mqtt.MessageHandler{
		"iotcihuy/home/temperature":    h.handleTemperature,
		"iotcihuy/home/humidity":       h.handleHumidity,
		"iotcihuy/home/gas":            h.handleGas,
		"iotcihuy/home/light":          h.handleLight,
		"iotcihuy/home/lamp/status":    h.handleLampStatus,
		"iotcihuy/home/door/status":    h.handleDoorStatus,
		"iotcihuy/home/door/verify":    h.handlePinVerification,
		"iotcihuy/home/curtain/status": h.handleCurtainStatus,
		"iotcihuy/home/debug":          h.handleDebug,
	}

	log.Printf("[MQTT] Connecting to broker... (Client connected: %v)", client.IsConnected())

	for topic, handler := range topics {
		if token := client.Subscribe(topic, 0, handler); token.Wait() && token.Error() != nil {
			log.Printf("[MQTT] Subscribe failed: %s | Error: %v", topic, token.Error())
		} else {
			log.Printf("[MQTT] Subscribed: %s", topic)
		}
	}
}

func (h *MQTTHandler) flushSensorCache() {
	h.sensorCache.mu.Lock()
	temperature := h.sensorCache.temperature
	temperaturePending := h.sensorCache.temperaturePending
	h.sensorCache.temperaturePending = false
	humidity := h.sensorCache.humidity
	humidityPending := h.sensorCache.humidityPending
	h.sensorCache.humidityPending = false
	light := h.sensorCache.light
	lightPending := h.sensorCache.lightPending
	h.sensorCache.lightPending = false
	gas := h.sensorCache.gas
	gasPending := h.sensorCache.gasPending
	h.sensorCache.gasPending = false
	h.sensorCache.mu.Unlock()

	if temperaturePending {
		if err := h.tempSvc.ProcessTemp(temperature); err != nil {
			log.Printf("[ERROR] Batch save temperature failed: %v", err)
		} else {
			log.Printf("[DEBUG] Batch saved temperature: %.2f", temperature)
		}
	}

	if humidityPending {
		if err := h.humidSvc.ProcessHumid(humidity); err != nil {
			log.Printf("[ERROR] Batch save humidity failed: %v", err)
		} else {
			log.Printf("[DEBUG] Batch saved humidity: %.2f", humidity)
		}
	}

	if lightPending {
		if err := h.lightSvc.ProcessLight(light); err != nil {
			log.Printf("[ERROR] Batch save light failed: %v", err)
		} else {
			log.Printf("[DEBUG] Batch saved light: %d Lux", light)
		}
	}

	if gasPending {
		if _, err := h.gasSvc.ProcessGas(gas); err != nil {
			log.Printf("[ERROR] Batch save gas failed: %v", err)
		} else {
			log.Printf("[DEBUG] Batch saved gas: %d PPM", gas)
		}
	}
}

// ==================== CONTROL FUNCTIONS (OUTPUT) ====================

func (h *MQTTHandler) PublishDoorControl(action string) error {
	topic := "iotcihuy/home/door/control"
	payload := map[string]string{
		"action": action,
		"method": "remote",
	}
	jsonPayload, _ := json.Marshal(payload)
	log.Printf("[MQTT] Publishing to %s: action=%s", topic, action)

	token := h.client.Publish(topic, 1, false, jsonPayload)
	token.Wait()

	if token.Error() != nil {
		log.Printf("[MQTT] Publish failed: %s | Error: %v", topic, token.Error())
	} else {
		log.Printf("[MQTT] Published: Door control: %s", action)
	}
	return token.Error()
}

func (h *MQTTHandler) PublishBuzzerControl(action string) {
	topic := "iotcihuy/home/buzzer/control"
	payload := map[string]string{
		"action": action,
		"source": "auto_alert",
	}
	jsonPayload, _ := json.Marshal(payload)
	log.Printf("[MQTT] Publishing to %s: action=%s", topic, action)

	token := h.client.Publish(topic, 0, false, jsonPayload)
	token.Wait()

	if token.Error() == nil {
		log.Printf("[MQTT] Published: Buzzer %s", action)
	} else {
		log.Printf("[MQTT] Buzzer publish failed: %v", token.Error())
	}
}

func (h *MQTTHandler) PublishLampControl(action string) {
	topic := "iotcihuy/home/lamp/control"

	payload := map[string]string{
		"action": action,
		"mode":   "auto",
	}
	jsonPayload, _ := json.Marshal(payload)
	log.Printf("[MQTT] Publishing to %s: action=%s, mode=auto", topic, action)

	token := h.client.Publish(topic, 1, false, jsonPayload)
	token.Wait()

	if token.Error() == nil {
		log.Printf("[MQTT] Published: Lamp %s (auto mode)", action)
	} else {
		log.Printf("[MQTT] Lamp publish failed: %v", token.Error())
	}
}

func (h *MQTTHandler) PublishCurtainControl(action string) error {
	topic := "iotcihuy/home/curtain/control"
	payload := map[string]string{
		"action": action,
	}
	jsonPayload, _ := json.Marshal(payload)
	log.Printf("[MQTT] Publishing to %s: action=%s", topic, action)

	token := h.client.Publish(topic, 1, false, jsonPayload)
	token.Wait()

	if token.Error() == nil {
		log.Printf("[MQTT] Published: Curtain -> %s", action)
	} else {
		log.Printf("[MQTT] Curtain publish failed: %v", token.Error())
	}

	return token.Error()
}

// ==================== SENSOR HANDLERS (INPUT) ====================

func (h *MQTTHandler) handleLight(client mqtt.Client, msg mqtt.Message) {
	var data struct {
		Lux  int    `json:"lux"`
		Unit string `json:"unit"`
	}
	if err := json.Unmarshal(msg.Payload(), &data); err != nil {
		return
	}

	log.Printf("Light: %d Lux", data.Lux)

	// Cache for batch persistence
	h.setLatestLight(data.Lux)

	h.wsHub.BroadcastData(msg.Payload())
}

// IMPROVED GAS HANDLER WITH MOVING AVERAGE
func (h *MQTTHandler) handleGas(client mqtt.Client, msg mqtt.Message) {
	var data struct {
		PPM  int    `json:"gas_ppm"`
		Unit string `json:"unit"`
	}
	if err := json.Unmarshal(msg.Payload(), &data); err != nil {
		log.Printf("[ERROR] JSON Parse Gas Failed: %v | Payload: %s", err, string(msg.Payload()))
		return
	}

	log.Printf("Gas: %d PPM (raw)", data.PPM)

	status := "normal"
	if data.PPM > 200 {
		status = "warning"
	}
	if data.PPM > 500 {
		status = "danger"
	}

	// Danger langsung disimpan; lainnya dibatch (disimpan saat flush 1 menit)
	if status == "danger" {
		go func(ppm int) {
			if savedStatus, err := h.gasSvc.ProcessGas(ppm); err != nil {
				log.Printf("[ERROR] Gas save failed: %v", err)
			} else {
				log.Printf("[DEBUG] Gas saved immediate: %d PPM (status=%s)", ppm, savedStatus)
			}
		}(data.PPM)
	} else {
		h.setLatestGas(data.PPM)
	}

	h.wsHub.BroadcastData(msg.Payload())
}

func (h *MQTTHandler) handleTemperature(client mqtt.Client, msg mqtt.Message) {
	log.Printf("[MQTT] Received on %s: %s", msg.Topic(), string(msg.Payload()))

	var data struct {
		Temperature float64 `json:"temperature"`
		Unit        string  `json:"unit"`
	}
	if err := json.Unmarshal(msg.Payload(), &data); err != nil {
		log.Printf("[MQTT] JSON Parse Temp Failed: %v | Payload: %s", err, string(msg.Payload()))
		return
	}

	log.Printf("[MQTT] Temperature: %.1f°C", data.Temperature)
	h.setLatestTemperature(data.Temperature)
	h.wsHub.BroadcastData(msg.Payload())
}

func (h *MQTTHandler) handleHumidity(client mqtt.Client, msg mqtt.Message) {
	var data struct {
		Humidity float64 `json:"humidity"`
		Unit     string  `json:"unit"`
	}
	if err := json.Unmarshal(msg.Payload(), &data); err != nil {
		log.Printf("[ERROR] JSON Parse Humid Failed: %v | Payload: %s", err, string(msg.Payload()))
		return
	}

	log.Printf("Humidity: %.1f%%", data.Humidity)
	h.setLatestHumidity(data.Humidity)
	h.wsHub.BroadcastData(msg.Payload())
}

// ==================== DEVICE STATUS HANDLERS ====================

func (h *MQTTHandler) handleLampStatus(client mqtt.Client, msg mqtt.Message) {
	var req struct {
		Status string `json:"status"`
		Mode   string `json:"mode"`
	}
	if err := json.Unmarshal(msg.Payload(), &req); err != nil {
		log.Printf("[ERROR] JSON Parse Lamp Failed: %v | Payload: %s", err, string(msg.Payload()))
		return
	}

	// Get previous mode from database
	lastLamp, err := h.lampSvc.GetLatest()
	previousMode := "auto"
	if err == nil {
		previousMode = lastLamp.Mode
	}

	// Snapshot old state before updating
	h.lampMutex.RLock()
	prevStatus := h.lastLampState
	h.lampMutex.RUnlock()

	// Check if mode changed
	modeChanged := previousMode != req.Mode
	statusChanged := prevStatus != req.Status

	// Update local state
	h.lampMutex.Lock()
	h.lastLampState = req.Status
	h.lastLampMode = req.Mode
	h.lastLampChangeTime = time.Now()
	h.lampMutex.Unlock()

	// Reset gas buffer when lamp changes (avoid initial spike)
	h.gasReadingMutex.Lock()
	h.gasReadings = make([]int, 0, h.maxGasReadings)
	h.gasReadingMutex.Unlock()

	// FORCE publish gas = 0 when lamp ON (EMI noise prevention)
	if req.Status == "on" {
		gasForcedZero := map[string]interface{}{
			"type":     "sensor_update",
			"sensor":   "gas",
			"ppm":      0,
			"unit":     "PPM",
			"enforced": "lamp_stabilization",
		}
		jsonData, _ := json.Marshal(gasForcedZero)
		h.wsHub.BroadcastData(jsonData)
		log.Printf("Gas: FORCED 0 PPM (Lamp ON - stabilization period)")
	}

	// Save to database when there is any change (status or mode)
	if modeChanged || statusChanged {
		go h.lampSvc.ProcessLamp(req.Status, req.Mode)
		log.Printf("Lamp: %s (mode: %s)", req.Status, req.Mode)
	}

	wsData := map[string]interface{}{
		"type":   "device_update",
		"device": "lamp",
		"state":  req.Status == "on",
		"mode":   req.Mode,
	}
	jsonData, _ := json.Marshal(wsData)
	h.wsHub.BroadcastData(jsonData)
}

func (h *MQTTHandler) handleDoorStatus(client mqtt.Client, msg mqtt.Message) {
	var req struct {
		Status string `json:"status"`
		Method string `json:"method"`
	}
	if err := json.Unmarshal(msg.Payload(), &req); err != nil {
		log.Printf("[ERROR] JSON Parse Door Failed: %v | Payload: %s", err, string(msg.Payload()))
		return
	}

	// Normalize method names
	if req.Method == "keypad" {
		req.Method = "pin"
	}
	if req.Method == "app_button" {
		req.Method = "remote"
	}

	go h.doorSvc.ProcessDoor(req.Status, req.Method, nil)

	if req.Status == "unlocked" {
		log.Printf("Door Access: %s", req.Method)
	}

	wsData := map[string]interface{}{
		"type":   "device_update",
		"device": "door",
		"locked": req.Status == "locked",
		"method": req.Method,
	}
	jsonData, _ := json.Marshal(wsData)
	h.wsHub.BroadcastData(jsonData)
}

func (h *MQTTHandler) handleCurtainStatus(client mqtt.Client, msg mqtt.Message) {
	var req struct {
		Status   string `json:"status"`
		Mode     string `json:"mode"`
		Position int    `json:"position"`
	}
	if err := json.Unmarshal(msg.Payload(), &req); err != nil {
		log.Printf("[ERROR] JSON Parse Curtain Failed: %v | Payload: %s", err, string(msg.Payload()))
		return
	}

	// Update in-memory state and detect changes without DB dependency
	h.curtainMutex.Lock()
	prevStatus := h.lastCurtainState
	prevMode := h.lastCurtainMode
	h.lastCurtainState = req.Status
	h.lastCurtainMode = req.Mode
	h.lastCurtainChangeTime = time.Now()
	h.curtainMutex.Unlock()

	statusChanged := prevStatus != req.Status
	modeChanged := prevMode != req.Mode

	go h.curtainSvc.ProcessCurtain(req.Status, req.Mode)

	// Only log if something actually changed
	if statusChanged || modeChanged {
		log.Printf("Curtain: %s (mode: %s)", req.Status, req.Mode)
	}

	wsData := map[string]interface{}{
		"type":   "device_update",
		"device": "curtain",
		"status": req.Status,
		"mode":   req.Mode,
	}
	jsonData, _ := json.Marshal(wsData)
	h.wsHub.BroadcastData(jsonData)
}

func (h *MQTTHandler) handleDebug(client mqtt.Client, msg mqtt.Message) {
	// Debug telemetry disabled for cleaner output
	// Data still received and processed by WebSocket
}

// ==================== PIN VERIFICATION HANDLER ====================

// handlePinVerification -
func (h *MQTTHandler) handlePinVerification(client mqtt.Client, msg mqtt.Message) {
	var req struct {
		Pin string `json:"pin"`
	}

	if err := json.Unmarshal(msg.Payload(), &req); err != nil {
		h.publishPinVerificationResponse(false, "Invalid format")
		return
	}

	// Get universal PIN from database
	pinData, err := h.pinSvc.GetUniversalPin()
	if err != nil {
		log.Printf("[ERROR] Failed to retrieve PIN from database: %v", err)
		h.publishPinVerificationResponse(false, "Database error")
		return
	}

	// Verify PIN
	if req.Pin != pinData.UniversalPin {
		log.Println("Invalid PIN")
		h.publishPinVerificationResponse(false, "Invalid PIN")
		return
	}

	log.Println("Valid PIN")

	// Send unlock command via MQTT
	h.PublishDoorControl("unlock")

	// Save access log to database (async)
	go h.doorSvc.ProcessDoor("unlocked", "pin", nil)

	// Send success response to ESP32
	h.publishPinVerificationResponse(true, "PIN verified, door unlocked")
}

// publishPinVerificationResponse - Send PIN verification result back to ESP32
func (h *MQTTHandler) publishPinVerificationResponse(valid bool, message string) {
	topic := "iotcihuy/home/door/verify/response"

	payload := map[string]interface{}{
		"valid":   valid,
		"message": message,
	}

	jsonPayload, _ := json.Marshal(payload)

	token := h.client.Publish(topic, 1, false, jsonPayload)
	token.Wait()
}
