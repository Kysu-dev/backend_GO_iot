package main

import (
	"log"
	"smarthome-backend/config"
	"smarthome-backend/internal/router"
)

func main() {
	config.InitDB() // Koneksi ke database
	r := router.InitRouter()

	log.Println("🚀 Server berjalan di http://localhost:8080")
	r.Run(":8080")
}
