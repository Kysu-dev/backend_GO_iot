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

	mqttLib "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	log.Println("╔════════════════════════════════════════╗")
	log.Println("║       Smart Home IoT Backend           ║")
	log.Println("╚════════════════════════════════════════╝")

	// 1. Load Config & DB
	cfg := config.LoadConfig()
	db := config.InitDB()

	// 2. Init Repositories
	gasRepo := repository.NewGasRepository(db)
	tempRepo := repository.NewTempRepository(db)
	humidRepo := repository.NewHumidRepository(db)
	lightRepo := repository.NewLightRepository(db)
	doorRepo := repository.NewDoorRepository(db)
	lampRepo := repository.NewLampRepository(db)
	curtainRepo := repository.NewCurtainRepository(db)
	userRepo := repository.NewUserRepository(db)
	accessLogRepo := repository.NewAccessLogRepository(db)
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
	pinSvc := service.NewPinService(pinRepo)
	sensorAnalyticsSvc := service.NewSensorAnalyticsService(db)

	// =================================================================
	// [KEMBALI KE LAMA] Hardcode URL & Secret (Supaya tidak Error Config)
	// =================================================================
	authSvc := service.NewAuthService("http://192.168.1.113:5001", "jwt-secret-key")

	// ================= SETUP MQTT (HiveMQ) =================
	opts := mqttLib.NewClientOptions()

	// Mengambil Broker dari .env (tcp://broker.hivemq.com:1883)
	// Jika error "cfg.MQTTBroker undefined", ganti baris ini dengan string langsung:
	// opts.AddBroker("tcp://broker.hivemq.com:1883")
	opts.AddBroker(cfg.MQTTBroker) 
	
	// ID Unik Random
	randomID := fmt.Sprintf("backend_%d", time.Now().UnixNano())
	opts.SetClientID(randomID)

	opts.SetCleanSession(true)
	opts.SetAutoReconnect(true)

	// Settingan Anti-Putus
	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(10 * time.Second)
	opts.SetWriteTimeout(10 * time.Second)

	opts.OnConnectionLost = func(c mqttLib.Client, err error) {
		log.Printf("[MQTT] Connection Lost: %v", err)
	}
	opts.OnConnect = func(c mqttLib.Client) {
		log.Println("[MQTT] Connected successfully to HiveMQ!")
	}

	mqttClient := mqttLib.NewClient(opts)

	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("[MQTT] Connection Failed:", token.Error())
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
	)

	// 7. Setup Routes
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
	authHandler := handler.NewAuthHandler(userSvc, authSvc)
	adminHandler := handler.NewAdminHandler(pinSvc, userSvc)

	deviceControlHandler := handler.NewDeviceControlHandler(
		mqttClient,
		lampSvc,
		doorSvc,
		curtainSvc,
	)

	faceHandler := handler.NewFaceHandler(accessLogSvc, mqttClient)
	sensorAnalyticsHandler := handler.NewSensorAnalyticsHandler(sensorAnalyticsSvc)
	dashboardHandler := handler.NewDashboardHandler(tempSvc, humidSvc, gasSvc, lightSvc, lampSvc, doorSvc, curtainSvc)

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
		AuthHandler:            authHandler,
		AdminHandler:           adminHandler,
		DeviceControlHandler:   deviceControlHandler,
		FaceHandler:            faceHandler,
		DashboardHandler:       dashboardHandler,
	}
	r := router.InitRouter(routerCfg)

	// 10. Run Server
	log.Println("╔════════════════════════════════════════╗")
	log.Printf("║   Server Running: http://localhost:%s   ║", cfg.ServerPort)
	log.Println("╚════════════════════════════════════════╝")

	r.Run(":" + cfg.ServerPort)
}