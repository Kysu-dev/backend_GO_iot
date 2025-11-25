package handler

import (
	"smarthome-backend/database/models"
	"smarthome-backend/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AccessLogHandler struct {
	svc service.AccessLogService
}

func NewAccessLogHandler(s service.AccessLogService) *AccessLogHandler {
	return &AccessLogHandler{svc: s}
}

func (h *AccessLogHandler) Create(c *gin.Context) {
	var req models.AccessLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	err := h.svc.LogAccess(req)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "Failed to log access"})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "Access logged successfully"})
}

func (h *AccessLogHandler) GetAll(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	logs, err := h.svc.GetAll(limit)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "Failed to retrieve logs"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": logs})
}

func (h *AccessLogHandler) GetByUserID(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "Invalid user ID"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	logs, err := h.svc.GetByUserID(uint(userID), limit)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "Failed to retrieve logs"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": logs})
}

func (h *AccessLogHandler) GetByStatus(c *gin.Context) {
	status := c.Param("status")
	if status != "success" && status != "failed" {
		c.JSON(400, gin.H{"success": false, "error": "Invalid status"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	logs, err := h.svc.GetByStatus(status, limit)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "Failed to retrieve logs"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": logs})
}
