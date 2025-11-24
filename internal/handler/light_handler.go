package handler

import (
	"smarthome-backend/database/models"
	"smarthome-backend/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type LightHandler struct {
	svc service.LightService
}

func NewLightHandler(s service.LightService) *LightHandler {
	return &LightHandler{svc: s}
}

func (h *LightHandler) Create(c *gin.Context) {
	var req models.LightRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	h.svc.ProcessLight(req.Lux)
	c.JSON(200, gin.H{"message": "Data saved"})
}

func (h *LightHandler) GetAll(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	data, _ := h.svc.GetHistory(limit)
	c.JSON(200, gin.H{"data": data})
}
