package handler

import (
	"fmt"
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
		c.JSON(500, gin.H{"success": false, "error": "Failed to retrieve universal PIN"})
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

	// ⭐ FIX: Pass 2 parameter dengan field yang benar
	if err := h.pinSvc.SetUniversalPin(req.UniversalPin, req.SetBy); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "Failed to set universal PIN"})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "Universal PIN updated successfully"})
}

func (h *AdminHandler) GetPendingUsers(c *gin.Context) {
	users, err := h.userSvc.GetPending()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "Failed to retrieve pending users"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": users})
}

// ⭐ Approve User
func (h *AdminHandler) Approve(c *gin.Context) {
	userID := c.Param("id")

	// Convert string to uint
	var id uint
	if _, err := fmt.Sscanf(userID, "%d", &id); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "Invalid user ID"})
		return
	}

	// Approve user
	if err := h.userSvc.Approve(id); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "Failed to approve user: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "User approved successfully",
		"user_id": id,
	})
}

// ⭐ Reject User
func (h *AdminHandler) Reject(c *gin.Context) {
	userID := c.Param("id")

	// Convert string to uint
	var id uint
	if _, err := fmt.Sscanf(userID, "%d", &id); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "Invalid user ID"})
		return
	}

	// Reject user (delete)
	if err := h.userSvc.Reject(id); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "Failed to reject user: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "User rejected successfully",
		"user_id": id,
	})
}

// ==================== ADMIN CRUD ====================

func (h *AdminHandler) CreateAdmin(c *gin.Context) {
	var req models.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	admin, err := h.userSvc.CreateAdmin(req)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"success": true, "data": admin})
}

func (h *AdminHandler) ListAdmins(c *gin.Context) {
	admins, err := h.userSvc.GetAdmins()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true, "data": admins})
}

func (h *AdminHandler) GetAdminByID(c *gin.Context) {
	var id uint
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &id); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "Invalid admin ID"})
		return
	}

	admin, err := h.userSvc.GetAdminByID(id)
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": admin})
}

func (h *AdminHandler) UpdateAdmin(c *gin.Context) {
	var id uint
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &id); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "Invalid admin ID"})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Force role admin regardless of payload
	req.Role = "admin"

	admin, err := h.userSvc.UpdateAdmin(id, req)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": admin})
}

func (h *AdminHandler) DeleteAdmin(c *gin.Context) {
	var id uint
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &id); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "Invalid admin ID"})
		return
	}

	if err := h.userSvc.DeleteAdmin(id); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "Admin deleted successfully"})
}
