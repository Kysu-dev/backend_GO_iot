package handler

import (
	"smarthome-backend/database/models"
	"smarthome-backend/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	svc service.NotificationService
}

func NewNotificationHandler(s service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: s}
}

func (h *NotificationHandler) Create(c *gin.Context) {
	var req models.NotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	err := h.svc.Create(req)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "Failed to create notification"})
		return
	}

	c.JSON(201, gin.H{"success": true, "message": "Notification created"})
}

func (h *NotificationHandler) GetAll(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	notifications, err := h.svc.GetAll(limit)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "Failed to retrieve notifications"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": notifications})
}

func (h *NotificationHandler) GetByType(c *gin.Context) {
	notifType := c.Param("type")
	validTypes := map[string]bool{"gas": true, "door": true, "system": true, "intruder": true}

	if !validTypes[notifType] {
		c.JSON(400, gin.H{"success": false, "error": "Invalid notification type"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	notifications, err := h.svc.GetByType(notifType, limit)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "Failed to retrieve notifications"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": notifications})
}
