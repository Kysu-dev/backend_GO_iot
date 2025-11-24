package handler

import (
	"strconv"
	"smarthome-backend/database/models"
	"smarthome-backend/internal/service"
	"github.com/gin-gonic/gin"
)

type GasHandler struct {
	svc service.GasService
}

func NewGasHandler(s service.GasService) *GasHandler {
	return &GasHandler{svc: s}
}

func (h *GasHandler) Create(c *gin.Context) {
	var req models.GasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	h.svc.ProcessGas(req.PPM)
	c.JSON(200, gin.H{"message": "Data saved"})
}

func (h *GasHandler) GetAll(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	data, _ := h.svc.GetHistory(limit)
	c.JSON(200, gin.H{"data": data})
}