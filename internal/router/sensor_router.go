package router

import (
	"smarthome-backend/internal/handler"
	"smarthome-backend/internal/websocket"
	"time"

	"github.com/gin-contrib/cors"
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
	AuthHandler      *handler.AuthHandler
	AdminHandler     *handler.AdminHandler

	// Notification Handler
	NotificationHandler *handler.NotificationHandler

	// Device Control Handler
	DeviceControlHandler *handler.DeviceControlHandler

	// Face Recognition Handler
	FaceHandler *handler.FaceHandler

	// WebSocket Hub
	WsHub *websocket.Hub
}

func InitRouter(cfg AppConfig) *gin.Engine {
	r := gin.Default()

	// ==================== NGROK-FRIENDLY CORS CONFIG ====================
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
			"Accept",
			"X-Requested-With",
			"ngrok-skip-browser-warning",
			"Content-Length",
			"Accept-Encoding",
			"X-CSRF-Token",
			"Cache-Control",
			"User-Agent",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"Content-Type",
			"ngrok-skip-browser-warning",
		},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// ==================== MANUAL CORS MIDDLEWARE (ENHANCED FOR NGROK) ====================
	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Allow all origins
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, ngrok-skip-browser-warning, User-Agent")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")

		// ⭐ Ngrok header passthrough
		c.Writer.Header().Set("ngrok-skip-browser-warning", "true")

		// ⭐ PREFLIGHT REQUEST HANDLING
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
			device.POST("/door/verify-pin", cfg.DoorHandler.VerifyPin)

			// Lamp Status
			device.POST("/lamp", cfg.LampHandler.Create)
			device.GET("/lamp/latest", cfg.LampHandler.GetLatest)
			device.GET("/lamp/history", cfg.LampHandler.GetAll)

			// Curtain Status
			device.POST("/curtain", cfg.CurtainHandler.Create)
			device.GET("/curtain/latest", cfg.CurtainHandler.GetLatest)
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

			// Manual Buzzer Control
			control.POST("/buzzer", cfg.DeviceControlHandler.ControlBuzzer)
		}

		// ==================== USER ENDPOINTS ====================
		user := api.Group("/user")
		{
			user.GET("/", cfg.UserHandler.GetAll)
			user.GET("/:id", cfg.UserHandler.GetByID)
			user.PUT("/:id", cfg.UserHandler.Update)
			user.DELETE("/:id", cfg.UserHandler.Delete)
		}

		// ==================== AUTH ENDPOINTS ====================
		auth := api.Group("/auth")
		{
			auth.POST("/register", cfg.AuthHandler.Register)
			auth.POST("/login", cfg.AuthHandler.Login)
		}

		// ==================== ADMIN ENDPOINTS ====================
		admin := api.Group("/admin")
		{
			// User Approval
			admin.GET("/users/pending", cfg.AdminHandler.GetPendingUsers)
			admin.POST("/users/:id/approve", cfg.AdminHandler.Approve)
			admin.POST("/users/:id/reject", cfg.AdminHandler.Reject)

			// Admin CRUD
			admin.POST("/admins", cfg.AdminHandler.CreateAdmin)
			admin.GET("/admins", cfg.AdminHandler.ListAdmins)
			admin.GET("/admins/:id", cfg.AdminHandler.GetAdminByID)
			admin.PUT("/admins/:id", cfg.AdminHandler.UpdateAdmin)
			admin.DELETE("/admins/:id", cfg.AdminHandler.DeleteAdmin)

			// Universal PIN Management
			admin.GET("/pin", cfg.AdminHandler.GetUniversalPin)
			admin.POST("/pin", cfg.AdminHandler.SetUniversalPin)
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

		// ==================== FACE RECOGNITION ENDPOINTS ====================
		face := api.Group("/face")
		{
			face.POST("/recognize", cfg.FaceHandler.RecognizeFace)
			face.POST("/enroll", cfg.FaceHandler.EnrollFace)
			face.POST("/reload", cfg.FaceHandler.ReloadFaces)
			face.GET("/logs", cfg.FaceHandler.GetAccessLogs)
		}

		// ==================== WEBSOCKET ENDPOINT ====================
		api.GET("/ws", cfg.WsHub.HandleWebSocket)
	}

	return r
}
