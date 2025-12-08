package handler

import (
	"smarthome-backend/database/models"
	"smarthome-backend/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type FaceHandler struct {
	svc service.FaceService
}

func NewFaceHandler(svc service.FaceService) *FaceHandler {
	return &FaceHandler{svc: svc}
}

// HandleRecognition - POST /api/face/recognition
// Receive recognition result from Python service
func (h *FaceHandler) HandleRecognition(c *gin.Context) {
	var req models.FaceRecognitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	if err := h.svc.ProcessRecognition(&req); err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "Recognition processed",
	})
}

// HandleAlert - POST /api/face/alert
// Receive alert from Python service
func (h *FaceHandler) HandleAlert(c *gin.Context) {
	var req models.FaceAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	if err := h.svc.ProcessAlert(&req); err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "Alert processed",
	})
}

// GetLogs - GET /api/face/logs
// Get recent recognition logs
func (h *FaceHandler) GetLogs(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	logs, err := h.svc.GetRecentLogs(limit)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    logs,
	})
}

// GetAlerts - GET /api/face/alerts
// Get unresolved alerts
func (h *FaceHandler) GetAlerts(c *gin.Context) {
	alerts, err := h.svc.GetUnresolvedAlerts()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    alerts,
	})
}

// ResolveAlert - POST /api/face/alerts/:id/resolve
// Resolve alert
func (h *FaceHandler) ResolveAlert(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "Invalid ID"})
		return
	}

	if err := h.svc.ResolveAlert(id); err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "Alert resolved",
	})
}
