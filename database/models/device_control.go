package models

// DeviceControl represents device control command via MQTT
type DeviceControl struct {
	Device string `json:"device" binding:"required,oneof=door lamp curtain"`
	Action string `json:"action" binding:"required"`
}
