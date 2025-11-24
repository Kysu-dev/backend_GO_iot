package router

import (
	"github.com/gin-gonic/gin"
	"smarthome-backend/internal/handler"
)

func InitRouter() *gin.Engine {
	r := gin.Default() // inisialisasi router Gin

	// Kelompokkan semua endpoint 
	api := r.Group("/api/sensor")
	{
		// Endpoint untuk sensor
		api.POST("/data", handler.LogEnvironmentData)  // kirim data sensor dari ESP32
		api.GET("/data/latest", handler.GetLatestData) // ambil data sensor terakhir
	}

	return r
}
