package handler

import (
	"smarthome-backend/database/models"
	"smarthome-backend/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type HumidHandler struct {
	svc service.HumidService
}

func NewHumidHandler(s service.HumidService) *HumidHandler {
	return &HumidHandler{svc: s}
}

func (h *HumidHandler) Create(c *gin.Context) {
	var req models.HumidRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	h.svc.ProcessHumid(req.Humidity)
	c.JSON(200, gin.H{"message": "Data saved"})
}

func (h *HumidHandler) GetAll(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	data, _ := h.svc.GetHistory(limit)
	c.JSON(200, gin.H{"data": data})
}
