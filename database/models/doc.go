// Package models contains all data models for Smart Home IoT system
//
// This package is organized by domain/functionality:
//
// Authentication & Users:
//   - user.go: User authentication and management models
//   - fingerprint.go: Fingerprint authentication models
//   - pin_code.go: PIN code authentication models
//   - access_log.go: Access history and logging models
//
// Camera & Vision:
//   - camera_capture.go: ESP32-CAM capture models
//
// Sensors:
//   - sensor_gas.go: Gas/smoke sensor models
//   - sensor_temperature.go: Temperature sensor models
//   - sensor_humidity.go: Humidity sensor models
//   - sensor_light.go: Light/LDR sensor models
//
// Actuators & Devices:
//   - door_status.go: Door lock control models
//   - lamp_status.go: Lamp control models
//   - curtain_status.go: Curtain/blind control models
//   - buzzer_log.go: Buzzer activity models
//
// System:
//   - notification.go: System notification models
//   - device_control.go: MQTT device control models
//   - response.go: Standard API response models
//
// All models use GORM for ORM and Gin validator for request validation.
package models
