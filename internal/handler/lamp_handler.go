package handler

import (
	"smarthome-backend/database/models"
	"smarthome-backend/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type LampHandler struct {
	svc service.LampService
}

func NewLampHandler(s service.LampService) *LampHandler {
	return &LampHandler{svc: s}
}

func (h *LampHandler) Create(c *gin.Context) {
	var req models.LampRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	err := h.svc.ProcessLamp(req.Status, req.Mode)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "Failed to save lamp status"})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "Lamp status saved"})
}

func (h *LampHandler) GetLatest(c *gin.Context) {
	data, err := h.svc.GetLatest()
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "No data found"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": data})
}

func (h *LampHandler) GetAll(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	data, err := h.svc.GetHistory(limit)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "Failed to retrieve data"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": data})
}
