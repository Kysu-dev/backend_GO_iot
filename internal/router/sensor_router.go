package router

import (
	"smarthome-backend/internal/handler"
	"smarthome-backend/internal/websocket"

	"github.com/gin-gonic/gin"
)

type AppConfig struct {
	GasHandler   *handler.GasHandler
	TempHandler  *handler.TempHandler
	HumidHandler *handler.HumidHandler
	LightHandler *handler.LightHandler
	WsHub        *websocket.Hub
}

func InitRouter(cfg AppConfig) *gin.Engine {
	r := gin.Default()

	api := r.Group("/api")
	{
		// Gas endpoints
		api.POST("/gas", cfg.GasHandler.Create)
		api.GET("/gas", cfg.GasHandler.GetAll)

		// Temperature endpoints
		api.POST("/temperature", cfg.TempHandler.Create)
		api.GET("/temperature", cfg.TempHandler.GetAll)

		// Humidity endpoints
		api.POST("/humidity", cfg.HumidHandler.Create)
		api.GET("/humidity", cfg.HumidHandler.GetAll)

		// Light endpoints
		api.POST("/light", cfg.LightHandler.Create)
		api.GET("/light", cfg.LightHandler.GetAll)

		// URL WebSocket
		api.GET("/ws", cfg.WsHub.HandleWebSocket)
	}
	return r
}
