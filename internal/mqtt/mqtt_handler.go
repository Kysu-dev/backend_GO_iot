package mqtt

import (
	"encoding/json"
	"log"
	"smarthome-backend/internal/service"
	"smarthome-backend/internal/websocket"

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
	wsHub      *websocket.Hub
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
	hub *websocket.Hub,
) *MQTTHandler {
	return &MQTTHandler{
		client:     client,
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
	topics := map[string]mqtt.MessageHandler{
		"iotcihuy/home/temperature":    h.handleTemperature,
		"iotcihuy/home/humidity":       h.handleHumidity,
		"iotcihuy/home/gas":            h.handleGas,
		"iotcihuy/home/light":          h.handleLight, // Logic Auto ada di sini
		"iotcihuy/home/lamp/status":    h.handleLampStatus,
		"iotcihuy/home/door/status":    h.handleDoorStatus,
		"iotcihuy/home/curtain/status": h.handleCurtainStatus,
	}

	for topic, handler := range topics {
		if token := client.Subscribe(topic, 0, handler); token.Wait() && token.Error() != nil {
			log.Printf("[ERROR] Subscribe Failed | Topic: %s | Error: %v", topic, token.Error())
		}
	}
	log.Println("[INFO] MQTT Handler Ready. Waiting for data...")
}

// ==================== CONTROL FUNCTIONS (OUTPUT) ====================

func (h *MQTTHandler) PublishDoorControl(action string) error {
	topic := "iotcihuy/home/door/control"
	payload := map[string]string{
		"action": action,
		"method": "remote",
	}
	jsonPayload, _ := json.Marshal(payload)
	token := h.client.Publish(topic, 1, false, jsonPayload)
	token.Wait()
	
	if token.Error() == nil {
		log.Printf("[OUTPUT] COMMAND DOOR | Action: %s", action)
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
	token := h.client.Publish(topic, 0, false, jsonPayload)
	token.Wait()

	if token.Error() != nil {
		log.Printf("[ERROR] Publish Buzzer Failed: %v", token.Error())
	} else {
		if action == "on" {
			log.Printf("[ALERT]  !!! DANGER DETECTED !!! Sending BUZZER ON.")
		}
	}
}

// --- FUNGSI PUBLISH LAMPU (WAJIB ADA UNTUK FITUR AUTO) ---
func (h *MQTTHandler) PublishLampControl(action string) {
	topic := "iotcihuy/home/lamp/control"
	
	// Payload mengirim mode: "auto" agar ESP32 tahu ini perintah otomatis
	payload := map[string]string{
		"action": action, // "on" atau "off"
		"mode":   "auto",
	}
	jsonPayload, _ := json.Marshal(payload)
	
	token := h.client.Publish(topic, 1, false, jsonPayload)
	token.Wait()
	
	if token.Error() != nil {
		log.Printf("[ERROR] Auto-Control Lamp Failed: %v", token.Error())
	} else {
		log.Printf("[AUTO]   SENSOR LOGIC | Lampu -> %s (Mode: Auto)", action)
	}
}

// ==================== SENSOR HANDLERS (INPUT) ====================

func (h *MQTTHandler) handleLight(client mqtt.Client, msg mqtt.Message) {
	var data struct {
		Lux  int    `json:"lux"`
		Unit string `json:"unit"`
	}
	if err := json.Unmarshal(msg.Payload(), &data); err != nil {
		log.Printf("[ERROR] JSON Parse Light Failed: %v", err)
		return
	}

	// 1. Simpan Data Lux (Sensor Log)
	go func() {
		if err := h.lightSvc.ProcessLight(data.Lux); err != nil {
			log.Printf("[ERROR] DB Save Light Failed: %v", err)
		}
	}()

	// 2. LOGIKA OTOMATISASI LAMPU (ANTI-TABRAKAN)
	go func() {
		// A. Cek Status Lampu Terakhir dari DB
		lastLamp, err := h.lampSvc.GetLatest()
		
		currentMode := "auto" // Default aman
		currentStatus := "off"

		if err == nil {
			currentMode = lastLamp.Mode
			currentStatus = lastLamp.Status
		}

		// B. CEK MODE: Jika MANUAL, Berhenti!
		if currentMode == "manual" {
			// User memegang kendali penuh. Sensor dilarang mengubah lampu.
			return 
		}

		// C. JIKA MODE AUTO: Lakukan Logika
		thresholdGelap := 300 // Angka batas gelap (Sesuaikan)

		// Kasus 1: Gelap DAN Lampu Mati -> NYALAKAN
		if data.Lux < thresholdGelap && currentStatus == "off" {
			h.PublishLampControl("on")
			h.lampSvc.ProcessLamp("on", "auto") // Update DB
		}

		// Kasus 2: Terang DAN Lampu Nyala -> MATIKAN
		if data.Lux > thresholdGelap && currentStatus == "on" {
			h.PublishLampControl("off")
			h.lampSvc.ProcessLamp("off", "auto") // Update DB
		}
	}()

	h.wsHub.BroadcastData(msg.Payload())
	log.Printf("[INPUT]  SENSOR LIGHT  | Value: %d Lux", data.Lux)
}

func (h *MQTTHandler) handleGas(client mqtt.Client, msg mqtt.Message) {
	var data struct {
		PPM  int    `json:"ppm"`
		Unit string `json:"unit"`
	}
	if err := json.Unmarshal(msg.Payload(), &data); err != nil {
		log.Printf("[ERROR] JSON Parse Gas Failed: %v", err)
		return
	}

	go func() {
		status, err := h.gasSvc.ProcessGas(data.PPM)
		if err != nil {
			log.Printf("[ERROR] DB Save Gas Failed: %v", err)
			return
		}

		if status == "warning" || status == "danger" {
			h.PublishBuzzerControl("on")
		} else {
			h.PublishBuzzerControl("off")
		}
	}()

	h.wsHub.BroadcastData(msg.Payload())
	log.Printf("[INPUT]  SENSOR GAS    | Value: %d PPM", data.PPM)
}

func (h *MQTTHandler) handleTemperature(client mqtt.Client, msg mqtt.Message) {
	var data struct {
		Temperature float64 `json:"temperature"`
		Unit        string  `json:"unit"`
	}
	if err := json.Unmarshal(msg.Payload(), &data); err != nil {
		log.Printf("[ERROR] JSON Parse Temp Failed: %v", err)
		return
	}
	
	go h.tempSvc.ProcessTemp(data.Temperature)
	h.wsHub.BroadcastData(msg.Payload())
	log.Printf("[INPUT]  SENSOR TEMP   | Value: %.1f C", data.Temperature)
}

func (h *MQTTHandler) handleHumidity(client mqtt.Client, msg mqtt.Message) {
	var data struct {
		Humidity float64 `json:"humidity"`
		Unit     string  `json:"unit"`
	}
	if err := json.Unmarshal(msg.Payload(), &data); err != nil {
		log.Printf("[ERROR] JSON Parse Humid Failed: %v", err)
		return
	}
	
	go h.humidSvc.ProcessHumid(data.Humidity)
	h.wsHub.BroadcastData(msg.Payload())
	log.Printf("[INPUT]  SENSOR HUMID  | Value: %.1f %%", data.Humidity)
}

// ==================== DEVICE STATUS HANDLERS ====================

func (h *MQTTHandler) handleLampStatus(client mqtt.Client, msg mqtt.Message) {
	var req struct {
		Status string `json:"status"`
		Mode   string `json:"mode"`
	}
	if err := json.Unmarshal(msg.Payload(), &req); err != nil {
		log.Printf("[ERROR] JSON Parse Lamp Failed: %v", err)
		return
	}
	
	// Simpan status balikan dari ESP32
	go h.lampSvc.ProcessLamp(req.Status, req.Mode)
	
	wsData := map[string]interface{}{
		"type": "device_update", "device": "lamp", 
		"state": req.Status == "on", "mode": req.Mode,
	}
	jsonData, _ := json.Marshal(wsData)
	h.wsHub.BroadcastData(jsonData)
	log.Printf("[INPUT]  DEVICE LAMP   | Status: %s | Mode: %s", req.Status, req.Mode)
}

func (h *MQTTHandler) handleDoorStatus(client mqtt.Client, msg mqtt.Message) {
	var req struct {
		Status string `json:"status"`
		Method string `json:"method"`
	}
	if err := json.Unmarshal(msg.Payload(), &req); err != nil {
		log.Printf("[ERROR] JSON Parse Door Failed: %v", err)
		return
	}
	
	// Auto Fix Method name for DB ENUM compatibility
	if req.Method == "keypad" { req.Method = "pin" }
	if req.Method == "app_button" { req.Method = "remote" }

	go h.doorSvc.ProcessDoor(req.Status, req.Method)
	
	wsData := map[string]interface{}{
		"type": "device_update", "device": "door", 
		"locked": req.Status == "locked", "method": req.Method,
	}
	jsonData, _ := json.Marshal(wsData)
	h.wsHub.BroadcastData(jsonData)
	log.Printf("[INPUT]  DEVICE DOOR   | Status: %s | Via: %s", req.Status, req.Method)
}

func (h *MQTTHandler) handleCurtainStatus(client mqtt.Client, msg mqtt.Message) {
	var req struct {
		Position int    `json:"position"`
		Mode     string `json:"mode"`
	}
	if err := json.Unmarshal(msg.Payload(), &req); err != nil {
		log.Printf("[ERROR] JSON Parse Curtain Failed: %v", err)
		return
	}
	
	go h.curtainSvc.ProcessCurtain(req.Position, req.Mode)
	
	wsData := map[string]interface{}{
		"type": "device_update", "device": "curtain", 
		"position": req.Position, "mode": req.Mode,
	}
	jsonData, _ := json.Marshal(wsData)
	h.wsHub.BroadcastData(jsonData)
	log.Printf("[INPUT]  DEVICE CURTAIN| Pos: %d%%", req.Position)
}