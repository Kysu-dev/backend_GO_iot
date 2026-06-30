package handler

import (
	"smarthome-backend/database/models"
	"smarthome-backend/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TempHandler struct {
	svc service.TempService
}

func NewTempHandler(s service.TempService) *TempHandler {
	return &TempHandler{svc: s}
}

func (h *TempHandler) Create(c *gin.Context) {
	var req models.TempRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	h.svc.ProcessTemp(req.Temperature)
	c.JSON(200, gin.H{"message": "Data saved"})
}

func (h *TempHandler) GetAll(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "1"))
	data, _ := h.svc.GetHistory(limit)
	c.JSON(200, gin.H{"data": data})
}
