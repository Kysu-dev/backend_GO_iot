package handler

import (
	"smarthome-backend/database/models"
	"smarthome-backend/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	svc service.UserService
}

func NewUserHandler(s service.UserService) *UserHandler {
	return &UserHandler{svc: s}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req models.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	user, err := h.svc.Register(req)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(201, gin.H{
		"success": true,
		"message": "User registered successfully",
		"data":    user,
	})
}

func (h *UserHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	user, err := h.svc.Login(req)
	if err != nil {
		c.JSON(401, gin.H{"success": false, "error": err.Error()})
		return
	}

	// TODO: Generate JWT token here
	// For now, just return user data
	c.JSON(200, gin.H{
		"success": true,
		"message": "Login successful",
		"data":    user,
		// "token": token,  // Add JWT token later
	})
}

func (h *UserHandler) GetAll(c *gin.Context) {
	users, err := h.svc.GetAll()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "Failed to retrieve users"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": users})
}

func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "Invalid user ID"})
		return
	}

	user, err := h.svc.GetByID(uint(id))
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "User not found"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": user})
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "Invalid user ID"})
		return
	}

	err = h.svc.Delete(uint(id))
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "Failed to delete user"})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "User deleted successfully"})
}
