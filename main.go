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
	log.Println("║   Smart Home IoT Backend Server      ║")
	log.Println("║         Starting...                   ║")
	log.Println("╚════════════════════════════════════════╝")

	// 1. Load Configuration
	cfg := config.LoadConfig()

	// 2. Init Database
	db := config.InitDB()

	// 3. Init MQTT Client
	opts := mqttLib.NewClientOptions()
	opts.AddBroker(cfg.MQTTBroker)
	opts.SetClientID("backend_service_" + fmt.Sprintf("%d", time.Now().Unix()))
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)

	mqttClient := mqttLib.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("❌ MQTT Connection Failed:", token.Error())
	}
	log.Println("✅ MQTT Connected!")

	// 4. Init WebSocket Hub
	wsHub := websocket.NewHub()
	go wsHub.Run()
	log.Println("✅ WebSocket Hub Running")

	// 5. Init Repositories
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

	// 6. Init Services
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

	// 7. Init Handlers
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
	deviceControlHandler := handler.NewDeviceControlHandler(mqttClient)
	faceHandler := handler.NewFaceHandler(accessLogSvc, mqttClient)

	// 8. Init MQTT Handler
	mqttH := mqtt.NewMQTTHandler(
		gasSvc,
		tempSvc,
		humidSvc,
		lightSvc,
		doorSvc,
		lampSvc,
		curtainSvc,
		wsHub,
	)
	mqttH.SetupRoutes(mqttClient)

	// 9. Router Configuration
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
