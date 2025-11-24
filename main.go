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
	// 1. Init Database (Dari Config)
	db := config.InitDB()

	// 2. Init MQTT Client (Langsung disini saja biar simple, atau buat file sendiri boleh)
	opts := mqttLib.NewClientOptions()
	opts.AddBroker("tcp://127.0.0.1:1883")
	opts.SetClientID("backend_service_" + fmt.Sprintf("%d", time.Now().Unix()))
	mqttClient := mqttLib.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("MQTT Fail:", token.Error())
	}
	log.Println("MQTT Connected!")

	// 3. Init WebSocket Hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// 4. Init Layers
	gasRepo := repository.NewGasRepository(db)
	gasSvc := service.NewGasService(gasRepo)
	gasHandler := handler.NewGasHandler(gasSvc)

	tempRepo := repository.NewTempRepository(db)
	tempSvc := service.NewTempService(tempRepo)
	tempHandler := handler.NewTempHandler(tempSvc)

	humidRepo := repository.NewHumidRepository(db)
	humidSvc := service.NewHumidService(humidRepo)
	humidHandler := handler.NewHumidHandler(humidSvc)

	lightRepo := repository.NewLightRepository(db)
	lightSvc := service.NewLightService(lightRepo)
	lightHandler := handler.NewLightHandler(lightSvc)

	// 5. Init MQTT Handler
	mqttH := mqtt.NewMQTTHandler(gasSvc, tempSvc, humidSvc, lightSvc, wsHub)
	mqttH.SetupRoutes(mqttClient)

	// 6. Router
	cfg := router.AppConfig{
		GasHandler:   gasHandler,
		TempHandler:  tempHandler,
		HumidHandler: humidHandler,
		LightHandler: lightHandler,
		WsHub:        wsHub,
	}
	r := router.InitRouter(cfg)

	// 7. Run
	r.Run(":8080")
}
