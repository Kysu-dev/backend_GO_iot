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

	// Buzzer state tracking
	lastBuzzerState string
	buzzerMutex     sync.Mutex

	// Lamp state tracking
	lastLampState      string
	lampMutex          sync.RWMutex
	lastLampChangeTime time.Time

	// Curtain state tracking
	lastCurtainState      string
	curtainMutex          sync.RWMutex
	lastCurtainChangeTime time.Time

	// ⭐ Gas moving average filter
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
	return &MQTTHandler{
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
		lastBuzzerState:    "off",
		lastLampState:      "off",
		lastLampChangeTime: time.Now(),
		gasReadings:        make([]int, 0, 5), // ⭐ Buffer 5 readings
		maxGasReadings:     5,
	}
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

	log.Printf("🔌 [MQTT] Connecting to broker... (Client connected: %v)", client.IsConnected())

	for topic, handler := range topics {
		if token := client.Subscribe(topic, 0, handler); token.Wait() && token.Error() != nil {
			log.Printf("❌ [MQTT] Subscribe failed: %s | Error: %v", topic, token.Error())
		} else {
			log.Printf("✅ [MQTT] Subscribed: %s", topic)
		}
	}
	log.Println("✅ [MQTT] All topics subscribed and ready")
}

// ==================== CONTROL FUNCTIONS (OUTPUT) ====================

func (h *MQTTHandler) PublishDoorControl(action string) error {
	topic := "iotcihuy/home/door/control"
	payload := map[string]string{
		"action": action,
		"method": "remote",
	}
	jsonPayload, _ := json.Marshal(payload)
	log.Printf("📤 [MQTT] Publishing to %s: action=%s", topic, action)

	token := h.client.Publish(topic, 1, false, jsonPayload)
	token.Wait()

	if token.Error() != nil {
		log.Printf("❌ [MQTT] Publish failed: %s | Error: %v", topic, token.Error())
	} else {
		log.Printf("✅ [MQTT] Published: 🚪 Door control: %s", action)
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
	log.Printf("📤 [MQTT] Publishing to %s: action=%s", topic, action)

	token := h.client.Publish(topic, 0, false, jsonPayload)
	token.Wait()

	if token.Error() == nil {
		log.Printf("✅ [MQTT] Published: 🚨 Buzzer %s", action)
	} else {
		log.Printf("❌ [MQTT] Buzzer publish failed: %v", token.Error())
	}
}

func (h *MQTTHandler) PublishLampControl(action string) {
	topic := "iotcihuy/home/lamp/control"

	payload := map[string]string{
		"action": action,
		"mode":   "auto",
	}
	jsonPayload, _ := json.Marshal(payload)
	log.Printf("📤 [MQTT] Publishing to %s: action=%s, mode=auto", topic, action)

	token := h.client.Publish(topic, 1, false, jsonPayload)
	token.Wait()

	if token.Error() == nil {
		log.Printf("✅ [MQTT] Published: 💡 Lamp %s (auto mode)", action)
	} else {
		log.Printf("❌ [MQTT] Lamp publish failed: %v", token.Error())
	}
}

func (h *MQTTHandler) PublishCurtainControl(action string) error {
	topic := "iotcihuy/home/curtain/control"
	payload := map[string]string{
		"action": action,
	}
	jsonPayload, _ := json.Marshal(payload)
	log.Printf("📤 [MQTT] Publishing to %s: action=%s", topic, action)

	token := h.client.Publish(topic, 1, false, jsonPayload)
	token.Wait()

	if token.Error() == nil {
		log.Printf("✅ [MQTT] Published: 🪟 Curtain → %s", action)
	} else {
		log.Printf("❌ [MQTT] Curtain publish failed: %v", token.Error())
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

	log.Printf("💡 Light: %d Lux", data.Lux)

	// Save to database
	go func() {
		if err := h.lightSvc.ProcessLight(data.Lux); err != nil {
			log.Printf("[ERROR] DB Save Light Failed: %v", err)
		}
	}()

	// Auto lamp control logic
	go func() {
		h.lampMutex.Lock()
		defer h.lampMutex.Unlock()

		currentStatus := h.lastLampState

		// Get current mode from database
		lastLamp, err := h.lampSvc.GetLatest()
		currentMode := "auto"
		if err == nil {
			currentMode = lastLamp.Mode
			if lastLamp.Status != currentStatus {
				currentStatus = lastLamp.Status
				h.lastLampState = currentStatus
			}
		}

		// Skip if manual mode
		if currentMode == "manual" {
			return
		}

		const (
			luxThreshold  = 500
			debounceDelay = 5 * time.Second
		)

		// Debouncing
		if time.Since(h.lastLampChangeTime) < debounceDelay {
			return
		}

		// Turn ON if dark (Lux < 500)
		if data.Lux < luxThreshold && currentStatus == "off" {
			h.PublishLampControl("on")
			h.lampSvc.ProcessLamp("on", "auto")
			h.lastLampState = "on"
			h.lastLampChangeTime = time.Now()
		}

		// Turn OFF if bright (Lux >= 500)
		if data.Lux >= luxThreshold && currentStatus == "on" {
			h.PublishLampControl("off")
			h.lampSvc.ProcessLamp("off", "auto")
			h.lastLampState = "off"
			h.lastLampChangeTime = time.Now()
		}
	}()

	// Auto curtain control logic
	go func() {
		h.curtainMutex.Lock()
		defer h.curtainMutex.Unlock()

		currentStatus := h.lastCurtainState

		// Get current mode from database
		lastCurtain, err := h.curtainSvc.GetLatest()
		currentMode := "auto"
		if err == nil {
			currentMode = lastCurtain.Mode
			if lastCurtain.Status != currentStatus {
				currentStatus = lastCurtain.Status
				h.lastCurtainState = currentStatus
			}
		}

		// Skip if manual mode
		if currentMode == "manual" {
			return
		}

		const (
			luxThreshold  = 500
			debounceDelay = 5 * time.Second
		)

		// Debouncing
		if time.Since(h.lastCurtainChangeTime) < debounceDelay {
			return
		}

		// CLOSE curtain if dark (Lux < 500) - Lamp is ON, keep gorden closed
		if data.Lux < luxThreshold && currentStatus == "open" {
			h.PublishCurtainControl("close")
			h.curtainSvc.ProcessCurtain("closed", "auto")
			h.lastCurtainState = "closed"
			h.lastCurtainChangeTime = time.Now()
		}

		// OPEN curtain if bright (Lux >= 500) - Use natural light
		if data.Lux >= luxThreshold && currentStatus == "closed" {
			h.PublishCurtainControl("open")
			h.curtainSvc.ProcessCurtain("open", "auto")
			h.lastCurtainState = "open"
			h.lastCurtainChangeTime = time.Now()
		}
	}()

	h.wsHub.BroadcastData(msg.Payload())
}

// ⭐ IMPROVED GAS HANDLER WITH MOVING AVERAGE
func (h *MQTTHandler) handleGas(client mqtt.Client, msg mqtt.Message) {
	var data struct {
		PPM  int    `json:"gas_ppm"`
		Unit string `json:"unit"`
	}
	if err := json.Unmarshal(msg.Payload(), &data); err != nil {
		log.Printf("[ERROR] JSON Parse Gas Failed: %v | Payload: %s", err, string(msg.Payload()))
		return
	}

	log.Printf("💨 Gas: %d PPM (Raw)", data.PPM)

	go func() {
		// Check if lamp just changed (stabilization period)
		h.lampMutex.RLock()
		timeSinceLampChange := time.Since(h.lastLampChangeTime)
		h.lampMutex.RUnlock()

		// Skip readings during lamp stabilization (10 seconds - same as ESP32 GAS_SKIP_DURATION)
		if timeSinceLampChange < 10*time.Second {
			log.Printf("⏭️  Gas: SKIPPED (Lamp stabilization: %.1fs remaining)", 10-(timeSinceLampChange.Seconds()))
			return
		}

		// ⭐ MOVING AVERAGE FILTER (reduces EMI noise)
		h.gasReadingMutex.Lock()
		h.gasReadings = append(h.gasReadings, data.PPM)
		if len(h.gasReadings) > h.maxGasReadings {
			h.gasReadings = h.gasReadings[1:] // Remove oldest reading
		}

		// Calculate average
		sum := 0
		for _, val := range h.gasReadings {
			sum += val
		}
		avgPPM := sum / len(h.gasReadings)
		h.gasReadingMutex.Unlock()

		// Save AVERAGED value to database (smoother data)
		_, err := h.gasSvc.ProcessGas(avgPPM)
		if err != nil {
			log.Printf("[ERROR] DB Save Gas Failed: %v", err)
			return
		}

		log.Printf("💨 Gas: %d PPM (Avg from %d readings)", avgPPM, len(h.gasReadings))

		// ⭐ Hysteresis threshold for buzzer (using averaged PPM)
		const (
			thresholdDanger = 800 // Turn buzzer ON
			thresholdSafe   = 500 // Turn buzzer OFF (prevent flapping)
		)

		h.buzzerMutex.Lock()
		defer h.buzzerMutex.Unlock()

		// Turn ON buzzer if gas high
		if avgPPM >= thresholdDanger && h.lastBuzzerState != "on" {
			h.PublishBuzzerControl("on")
			h.lastBuzzerState = "on"
		}

		// Turn OFF buzzer if gas low
		if avgPPM < thresholdSafe && h.lastBuzzerState == "on" {
			h.PublishBuzzerControl("off")
			h.lastBuzzerState = "off"
		}
	}()

	h.wsHub.BroadcastData(msg.Payload())
}

func (h *MQTTHandler) handleTemperature(client mqtt.Client, msg mqtt.Message) {
	log.Printf("📥 [MQTT] Received on %s: %s", msg.Topic(), string(msg.Payload()))

	var data struct {
		Temperature float64 `json:"temperature"`
		Unit        string  `json:"unit"`
	}
	if err := json.Unmarshal(msg.Payload(), &data); err != nil {
		log.Printf("❌ [MQTT] JSON Parse Temp Failed: %v | Payload: %s", err, string(msg.Payload()))
		return
	}

	log.Printf("✅ [MQTT] 🌡️  Temperature: %.1f°C", data.Temperature)
	go h.tempSvc.ProcessTemp(data.Temperature)
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

	log.Printf("💧 Humidity: %.1f%%", data.Humidity)
	go h.humidSvc.ProcessHumid(data.Humidity)
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

	// Check if mode changed
	modeChanged := previousMode != req.Mode

	// Update local state
	h.lampMutex.Lock()
	h.lastLampState = req.Status
	h.lastLampChangeTime = time.Now()
	h.lampMutex.Unlock()

	// ⭐ Reset gas buffer saat lamp berubah (menghindari spike awal)
	h.gasReadingMutex.Lock()
	h.gasReadings = make([]int, 0, h.maxGasReadings)
	h.gasReadingMutex.Unlock()

	// ⭐ FORCE publish gas = 0 saat lamp ON (EMI noise prevention)
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
		log.Printf("💨 Gas: FORCED 0 PPM (Lamp ON - stabilization period)")
	}

	// Only save to database if mode changed OR status changed in manual mode
	if modeChanged || req.Mode == "manual" {
		go h.lampSvc.ProcessLamp(req.Status, req.Mode)
		if modeChanged {
			log.Printf("💡 Lamp: %s (mode: %s)", req.Status, req.Mode)
		}
	}

	// Trigger auto control immediately if switched to auto mode
	if modeChanged && req.Mode == "auto" {
		go func() {
			time.Sleep(500 * time.Millisecond) // Small delay to ensure state is updated

			// Get latest light sensor value
			latestLight, err := h.lightSvc.GetLatest()
			if err != nil {
				log.Printf("[ERROR] Failed to get latest light value for auto mode: %v", err)
				return
			}

			const luxThreshold = 500

			h.lampMutex.Lock()
			currentStatus := h.lastLampState
			h.lampMutex.Unlock()

			// Apply auto logic based on current lux
			if latestLight.Lux < luxThreshold && currentStatus == "off" {
				h.PublishLampControl("on")
				h.lampSvc.ProcessLamp("on", "auto")
				h.lampMutex.Lock()
				h.lastLampState = "on"
				h.lastLampChangeTime = time.Now()
				h.lampMutex.Unlock()
			} else if latestLight.Lux >= luxThreshold && currentStatus == "on" {
				h.PublishLampControl("off")
				h.lampSvc.ProcessLamp("off", "auto")
				h.lampMutex.Lock()
				h.lastLampState = "off"
				h.lastLampChangeTime = time.Now()
				h.lampMutex.Unlock()
			}
		}()
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
		log.Printf("🚪 Access: %s", req.Method)
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

	// Get previous status to detect changes
	lastCurtain, _ := h.curtainSvc.GetLatest()
	statusChanged := lastCurtain == nil || lastCurtain.Status != req.Status
	modeChanged := lastCurtain == nil || lastCurtain.Mode != req.Mode

	go h.curtainSvc.ProcessCurtain(req.Status, req.Mode)

	// Only log if something actually changed
	if statusChanged || modeChanged {
		log.Printf("🪟 Curtain: %s (mode: %s)", req.Status, req.Mode)
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
		log.Println("❌ Invalid PIN")
		h.publishPinVerificationResponse(false, "Invalid PIN")
		return
	}

	log.Println("✅ Valid PIN")

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
