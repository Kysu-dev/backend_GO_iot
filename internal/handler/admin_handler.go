package handler

import (
	"smarthome-backend/database/models"
	"smarthome-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	pinSvc  service.PinService
	userSvc service.UserService
}

func NewAdminHandler(pSvc service.PinService, uSvc service.UserService) *AdminHandler {
	return &AdminHandler{
		pinSvc:  pSvc,
		userSvc: uSvc,
	}
}

func (h *AdminHandler) GetUniversalPin(c *gin.Context) {
	pin, err := h.pinSvc.GetUniversalPin()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "Failed to get PIN"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": pin})
}

func (h *AdminHandler) SetUniversalPin(c *gin.Context) {
	var req models.PinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	adminID := uint(1)

	err := h.pinSvc.SetUniversalPin(req.UniversalPin, adminID)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "Failed to update PIN"})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "PIN updated successfully"})
}

func (h *AdminHandler) GetPendingUsers(c *gin.Context) {
	users, err := h.userSvc.GetPending()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "Failed to retrieve pending users"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": users})
}
