package main

import (
	"fmt"
	"log"
	"time"

	"smarthome-backend/config"
	"smarthome-backend/internal/handler"
	"smarthome-backend/internal/mqtt"
	"smarthome-backend/internal/repository"
	"smarthome-backend/internal/router"
	"smarthome-backend/internal/service"
	"smarthome-backend/internal/websocket"

	mqttLib "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	log.Println("╔════════════════════════════════════════╗")
	log.Println("║   🏠 Smart Home IoT Backend           ║")
	log.Println("╚════════════════════════════════════════╝")

	// 1. Load Config & DB
	cfg := config.LoadConfig()
	db := config.InitDB()

	// 2. Init WebSocket
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// 3. Init Repositories
	gasRepo := repository.NewGasRepository(db)
	tempRepo := repository.NewTempRepository(db)
	humidRepo := repository.NewHumidRepository(db)
	lightRepo := repository.NewLightRepository(db)
	doorRepo := repository.NewDoorRepository(db)
	lampRepo := repository.NewLampRepository(db)
	curtainRepo := repository.NewCurtainRepository(db)
	userRepo := repository.NewUserRepository(db)
	accessLogRepo := repository.NewAccessLogRepository(db)
	notifRepo := repository.NewNotificationRepository(db)
	pinRepo := repository.NewPinRepository(db)

	// 4. Init Services
	gasSvc := service.NewGasService(gasRepo)
	tempSvc := service.NewTempService(tempRepo)
	humidSvc := service.NewHumidService(humidRepo)
	lightSvc := service.NewLightService(lightRepo)
	doorSvc := service.NewDoorService(doorRepo, accessLogRepo)
	lampSvc := service.NewLampService(lampRepo)
	curtainSvc := service.NewCurtainService(curtainRepo)
	userSvc := service.NewUserService(userRepo)
	accessLogSvc := service.NewAccessLogService(accessLogRepo)
	notifSvc := service.NewNotificationService(notifRepo)
	pinSvc := service.NewPinService(pinRepo)
	sensorAnalyticsSvc := service.NewSensorAnalyticsService(db)

	// Server Python Face Rec
	authSvc := service.NewAuthService("http://10.124.88.112:5001", "jwt-secret-key")

	// =========================================================================
	// 5. SETUP MQTT CLIENT (OPTIMALISASI HIVEMQ)
	// =========================================================================
	opts := mqttLib.NewClientOptions()
	opts.AddBroker(cfg.MQTTBroker) // Pastikan config isinya: tcp://broker.hivemq.com:1883

	// --- A. ID UNIK (PENTING) ---
	// Menggunakan Nano Second agar ID selalu beda tiap kali run.
	// Ini mencegah error "Connection Lost" karena rebutan ID.
	randomID := fmt.Sprintf("backend_%d", time.Now().UnixNano())
	opts.SetClientID(randomID)

	opts.SetCleanSession(true)
	opts.SetAutoReconnect(true)

	// --- B. SETTINGAN ANTI-PUTUS (PENTING UNTUK HIVEMQ) ---
	// HiveMQ publik kadang lambat responnya. Kita perpanjang waktu tunggunya.
	opts.SetKeepAlive(60 * time.Second)    // Kirim sinyal "Saya Hidup" tiap 60 detik
	opts.SetPingTimeout(10 * time.Second)  // Tunggu balasan server 10 detik (Default cuma 2s)
	opts.SetWriteTimeout(10 * time.Second) // Tunggu proses kirim data 10 detik
	// ------------------------------------------------------

	opts.OnConnectionLost = func(c mqttLib.Client, err error) {
		log.Println("⚠️ MQTT disconnected")
	}
	opts.OnConnect = func(c mqttLib.Client) {
		log.Println("✅ MQTT connected")
	}

	mqttClient := mqttLib.NewClient(opts)

	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("❌ MQTT Connection Failed:", token.Error())
	}

	// 6. Init MQTT Handler
	mqttH := mqtt.NewMQTTHandler(
		mqttClient,
		gasSvc,
		tempSvc,
		humidSvc,
		lightSvc,
		doorSvc,
		lampSvc,
		curtainSvc,
		pinSvc,
		wsHub,
	)

	// 7. Setup Routes (Subscribe Topik)
	mqttH.SetupRoutes(mqttClient)

	// 8. Init Handlers (HTTP)
	gasHandler := handler.NewGasHandler(gasSvc)
	tempHandler := handler.NewTempHandler(tempSvc)
	humidHandler := handler.NewHumidHandler(humidSvc)
	lightHandler := handler.NewLightHandler(lightSvc)
	doorHandler := handler.NewDoorHandler(doorSvc, pinSvc, mqttClient)
	lampHandler := handler.NewLampHandler(lampSvc)
	curtainHandler := handler.NewCurtainHandler(curtainSvc)
	userHandler := handler.NewUserHandler(userSvc)
	accessLogHandler := handler.NewAccessLogHandler(accessLogSvc)
	notifHandler := handler.NewNotificationHandler(notifSvc)
	authHandler := handler.NewAuthHandler(userSvc, authSvc)
	adminHandler := handler.NewAdminHandler(pinSvc, userSvc)

	// --- Device Handler (Lengkap 4 Parameter) ---
	deviceControlHandler := handler.NewDeviceControlHandler(
		mqttClient,
		lampSvc,
		doorSvc,
		curtainSvc,
	)

	faceHandler := handler.NewFaceHandler(accessLogSvc, mqttClient)
	sensorAnalyticsHandler := handler.NewSensorAnalyticsHandler(sensorAnalyticsSvc)

	// 9. Router Configuration
	routerCfg := router.AppConfig{
		GasHandler:             gasHandler,
		TempHandler:            tempHandler,
		HumidHandler:           humidHandler,
		LightHandler:           lightHandler,
		SensorAnalyticsHandler: sensorAnalyticsHandler,
		DoorHandler:            doorHandler,
		LampHandler:            lampHandler,
		CurtainHandler:         curtainHandler,
		UserHandler:            userHandler,
		AccessLogHandler:       accessLogHandler,
		NotificationHandler:    notifHandler,
		AuthHandler:            authHandler,
		AdminHandler:           adminHandler,
		DeviceControlHandler:   deviceControlHandler,
		FaceHandler:            faceHandler,
		WsHub:                  wsHub,
	}
	r := router.InitRouter(routerCfg)

	// 10. Run Server
	log.Println("╔════════════════════════════════════════╗")
	log.Printf("║  🚀 Server: http://localhost:%s       ║", cfg.ServerPort)
	log.Println("╚════════════════════════════════════════╝")

	r.Run(":" + cfg.ServerPort)
}
