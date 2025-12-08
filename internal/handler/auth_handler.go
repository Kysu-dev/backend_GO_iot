package handler

import (
	"smarthome-backend/database/models"
	"smarthome-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	userSvc service.UserService
	authSvc service.AuthService
}

func NewAuthHandler(uSvc service.UserService, aSvc service.AuthService) *AuthHandler {
	return &AuthHandler{
		userSvc: uSvc,
		authSvc: aSvc,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	if req.FaceImage != "" {
		valid, err := h.authSvc.ValidateFaceWithPython(req.FaceImage)
		if err != nil || !valid {
			c.JSON(400, gin.H{"success": false, "error": "No face detected or invalid image"})
			return
		}
	}

	user, err := h.userSvc.Register(req)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	if req.FaceImage != "" {
		filename, err := h.authSvc.EnrollFaceWithPython(user.UserID, user.Name, req.FaceImage)
		if err != nil {
			c.JSON(500, gin.H{"success": false, "error": "User created but face enrollment failed"})
			return
		}
		user.FaceEncodingPath = filename
		h.userSvc.Update(user)
	}

	c.JSON(201, gin.H{
		"success": true,
		"message": "Registration successful, waiting for admin approval",
		"data":    user,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	user, err := h.userSvc.Login(req)
	if err != nil {
		c.JSON(401, gin.H{"success": false, "error": err.Error()})
		return
	}

	token, _ := h.authSvc.GenerateToken(user)

	c.JSON(200, gin.H{
		"success": true,
		"message": "Login successful",
		"token":   token,
		"user":    user,
	})
}
