package router

import (
	"smarthome-backend/internal/handler"
	"smarthome-backend/internal/websocket"

	"github.com/gin-gonic/gin"
)

type AppConfig struct {
	// Sensor Handlers
	GasHandler   *handler.GasHandler
	TempHandler  *handler.TempHandler
	HumidHandler *handler.HumidHandler
	LightHandler *handler.LightHandler

	// Device Handlers
	DoorHandler    *handler.DoorHandler
	LampHandler    *handler.LampHandler
	CurtainHandler *handler.CurtainHandler

	// User & Auth Handlers
	UserHandler      *handler.UserHandler
	AccessLogHandler *handler.AccessLogHandler

	// Notification Handler
	NotificationHandler *handler.NotificationHandler

	// Device Control Handler
	DeviceControlHandler *handler.DeviceControlHandler

	// WebSocket Hub
	WsHub *websocket.Hub
}

func InitRouter(cfg AppConfig) *gin.Engine {
	r := gin.Default()

	// CORS Middleware (if needed)
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "smart-home-backend",
			"version": "1.0.0",
		})
	})

	// API Routes
	api := r.Group("/api")
	{
		// ==================== SENSOR ENDPOINTS ====================
		sensor := api.Group("/sensor")
		{
			// Gas Sensor
			sensor.POST("/gas", cfg.GasHandler.Create)
			sensor.GET("/gas", cfg.GasHandler.GetAll)

			// Temperature Sensor
			sensor.POST("/temperature", cfg.TempHandler.Create)
			sensor.GET("/temperature", cfg.TempHandler.GetAll)

			// Humidity Sensor
			sensor.POST("/humidity", cfg.HumidHandler.Create)
			sensor.GET("/humidity", cfg.HumidHandler.GetAll)

			// Light Sensor
			sensor.POST("/light", cfg.LightHandler.Create)
			sensor.GET("/light", cfg.LightHandler.GetAll)
		}

		// ==================== DEVICE STATUS ENDPOINTS ====================
		device := api.Group("/device")
		{
			// Door Status
			device.POST("/door", cfg.DoorHandler.Create)
			device.GET("/door/latest", cfg.DoorHandler.GetLatest)
			device.GET("/door/history", cfg.DoorHandler.GetAll)

			// Lamp Status
			device.POST("/lamp", cfg.LampHandler.Create)
			device.GET("/lamp/latest", cfg.LampHandler.GetLatest)
			device.GET("/lamp/history", cfg.LampHandler.GetAll)

			// Curtain Status
			device.POST("/curtain", cfg.CurtainHandler.Create)
			device.GET("/curtain/latest", cfg.CurtainHandler.GetLatest)
			device.GET("/curtain/history", cfg.CurtainHandler.GetAll)
		}

		// ==================== DEVICE CONTROL ENDPOINTS ====================
		control := api.Group("/control")
		{
			// Universal control endpoint
			control.POST("/device", cfg.DeviceControlHandler.Control)

			// Specific device controls
			control.POST("/door", cfg.DeviceControlHandler.ControlDoor)
			control.POST("/lamp", cfg.DeviceControlHandler.ControlLamp)
			control.POST("/curtain", cfg.DeviceControlHandler.ControlCurtain)
		}

		// ==================== USER ENDPOINTS ====================
		user := api.Group("/user")
		{
			user.POST("/register", cfg.UserHandler.Register)
			user.POST("/login", cfg.UserHandler.Login)
			user.GET("/", cfg.UserHandler.GetAll)
			user.GET("/:id", cfg.UserHandler.GetByID)
			user.DELETE("/:id", cfg.UserHandler.Delete)
		}

		// ==================== ACCESS LOG ENDPOINTS ====================
		accessLog := api.Group("/access-log")
		{
			accessLog.POST("/", cfg.AccessLogHandler.Create)
			accessLog.GET("/", cfg.AccessLogHandler.GetAll)
			accessLog.GET("/user/:user_id", cfg.AccessLogHandler.GetByUserID)
			accessLog.GET("/status/:status", cfg.AccessLogHandler.GetByStatus)
		}

		// ==================== NOTIFICATION ENDPOINTS ====================
		notification := api.Group("/notification")
		{
			notification.POST("/", cfg.NotificationHandler.Create)
			notification.GET("/", cfg.NotificationHandler.GetAll)
			notification.GET("/type/:type", cfg.NotificationHandler.GetByType)
		}

		// ==================== WEBSOCKET ENDPOINT ====================
		api.GET("/ws", cfg.WsHub.HandleWebSocket)
	}

	return r
}
