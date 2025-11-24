package config

import (
	"log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)


func InitDB() *gorm.DB {

	dsn := "root:@tcp(127.0.0.1:3306)/smarthome_db?charset=utf8mb4&parseTime=True&loc=Local"
	
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Gagal koneksi Database:", err)
	}

	log.Println("Database Connected!")
	
	return db // <--- WAJIB ADA INI
}