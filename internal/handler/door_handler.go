package handler

import (
	"smarthome-backend/database/models"
	"smarthome-backend/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DoorHandler struct {
	svc service.DoorService
}

func NewDoorHandler(s service.DoorService) *DoorHandler {
	return &DoorHandler{svc: s}
}

func (h *DoorHandler) Create(c *gin.Context) {
	var req models.DoorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	err := h.svc.ProcessDoor(req.Status, req.Method)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "Failed to save door status"})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "Door status saved"})
}

func (h *DoorHandler) GetLatest(c *gin.Context) {
	data, err := h.svc.GetLatest()
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "No data found"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": data})
}

func (h *DoorHandler) GetAll(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	data, err := h.svc.GetHistory(limit)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "Failed to retrieve data"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": data})
}
