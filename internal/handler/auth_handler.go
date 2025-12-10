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

	if req.FaceImage == "" {
		c.JSON(400, gin.H{"success": false, "error": "Face image is required"})
		return
	}

	// STEP 1: Validate face SEBELUM create user
	valid, err := h.authSvc.ValidateFaceWithPython(req.FaceImage)
	if err != nil || !valid {
		c.JSON(400, gin.H{"success": false, "error": "Face validation failed: " + err.Error()})
		return
	}

	// STEP 2: Create user (face sudah valid)
	user, err := h.userSvc.Register(req)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// STEP 3: Enroll face & dapat path pkl
	filename, err := h.authSvc.EnrollFaceWithPython(user.UserID, user.Name, req.FaceImage)
	if err != nil {
		h.userSvc.Delete(user.UserID) // Rollback
		c.JSON(500, gin.H{"success": false, "error": "Face enrollment failed: " + err.Error()})
		return
	}

	// STEP 4: Update user dengan face_encoding_path
	user.FaceEncodingPath = filename
	h.userSvc.Update(user)

	c.JSON(201, gin.H{
		"success": true,
		"message": "Registration successful. Waiting for admin approval.",
		"data": gin.H{
			"user_id":            user.UserID,
			"name":               user.Name,
			"email":              user.Email,
			"status":             user.Status,
			"face_encoding_path": user.FaceEncodingPath,
		},
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
