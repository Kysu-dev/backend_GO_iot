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
	log.Println("║   Smart Home IoT Backend Server        ║")
	log.Println("║           Starting...                  ║")
	log.Println("╚════════════════════════════════════════╝")

	// 1. Load Config & DB
	cfg := config.LoadConfig()
	db := config.InitDB()

	// 2. Init WebSocket
	wsHub := websocket.NewHub()
	go wsHub.Run()
	// go wsHub.StartPingTimer() // Aktifkan jika sudah implementasi ping
	log.Println("✅ WebSocket Hub Running")

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
	doorSvc := service.NewDoorService(doorRepo)
	lampSvc := service.NewLampService(lampRepo)
	curtainSvc := service.NewCurtainService(curtainRepo)
	userSvc := service.NewUserService(userRepo)
	accessLogSvc := service.NewAccessLogService(accessLogRepo)
	notifSvc := service.NewNotificationService(notifRepo)
	pinSvc := service.NewPinService(pinRepo)
	authSvc := service.NewAuthService("http://localhost:5000", "jwt-secret-key")

	// =========================================================================
	// 5. SETUP MQTT CLIENT (PINDAHKAN KE ATAS SINI AGAR BISA DIPAKAI HANDLER)
	// =========================================================================
	opts := mqttLib.NewClientOptions()
	opts.AddBroker(cfg.MQTTBroker)
	opts.SetClientID(cfg.MQTTClientID)
	// Jika config kosong, fallback ke random ID
	if cfg.MQTTClientID == "" {
		opts.SetClientID("backend_srv_" + fmt.Sprintf("%d", time.Now().Unix()))
	}
	opts.SetCleanSession(true)
	opts.SetAutoReconnect(true)

	// Buat object client (TAPI BELUM CONNECT)
	mqttClient := mqttLib.NewClient(opts)

	// 6. Init MQTT Handler (SEKARANG KITA BISA MASUKKAN mqttClient)
	mqttH := mqtt.NewMQTTHandler(
		mqttClient, // <--- PARAMETER INI YANG TADI KURANG!
		gasSvc,
		tempSvc,
		humidSvc,
		lightSvc,
		doorSvc,
		lampSvc,
		curtainSvc,
		wsHub,
	)

	// 7. Setup Callback & Connect
	// Kita set OnConnect handler setelah mqttH jadi, supaya bisa panggil SetupRoutes
	opts.OnConnect = func(c mqttLib.Client) {
		log.Println("✅ Connected to MQTT Broker")
		mqttH.SetupRoutes(c)
	}
	opts.OnConnectionLost = func(c mqttLib.Client, err error) {
		log.Printf("⚠️ MQTT Connection Lost: %v", err)
	}

	// Karena kita mengubah opts setelah NewClient, kita perlu init ulang client
	// ATAU cara paling gampang: Lakukan connect dan panggil SetupRoutes manual.
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("❌ MQTT Connection Failed:", token.Error())
	}
	// Panggil manual agar langsung subscribe saat start
	mqttH.SetupRoutes(mqttClient)

	// 8. Init Handlers (HTTP)
	gasHandler := handler.NewGasHandler(gasSvc)
	tempHandler := handler.NewTempHandler(tempSvc)
	humidHandler := handler.NewHumidHandler(humidSvc)
	lightHandler := handler.NewLightHandler(lightSvc)
	doorHandler := handler.NewDoorHandler(doorSvc)
	lampHandler := handler.NewLampHandler(lampSvc)
	curtainHandler := handler.NewCurtainHandler(curtainSvc)
	userHandler := handler.NewUserHandler(userSvc)
	accessLogHandler := handler.NewAccessLogHandler(accessLogSvc)
	notifHandler := handler.NewNotificationHandler(notifSvc)
	authHandler := handler.NewAuthHandler(userSvc, authSvc)
	adminHandler := handler.NewAdminHandler(pinSvc, userSvc)

	deviceControlHandler := handler.NewDeviceControlHandler(mqttClient)
	faceHandler := handler.NewFaceHandler(accessLogSvc, mqttClient) // 9. Router Configuration
	routerCfg := router.AppConfig{
		GasHandler:           gasHandler,
		TempHandler:          tempHandler,
		HumidHandler:         humidHandler,
		LightHandler:         lightHandler,
		DoorHandler:          doorHandler,
		LampHandler:          lampHandler,
		CurtainHandler:       curtainHandler,
		UserHandler:          userHandler,
		AccessLogHandler:     accessLogHandler,
		NotificationHandler:  notifHandler,
		AuthHandler:          authHandler,
		AdminHandler:         adminHandler,
		DeviceControlHandler: deviceControlHandler,
		FaceHandler:          faceHandler,
		WsHub:                wsHub,
	}
	r := router.InitRouter(routerCfg)

	// 10. Run Server
	log.Println("\n╔════════════════════════════════════════╗")
	log.Printf("║  🚀 Server running on port %s        ║\n", cfg.ServerPort)
	log.Println("║  📱 API: http://localhost:" + cfg.ServerPort + "      ║")
	log.Println("║  📡 MQTT: " + cfg.MQTTBroker + "     ║")
	log.Println("╚════════════════════════════════════════╝\n")

	r.Run(":" + cfg.ServerPort)
}
