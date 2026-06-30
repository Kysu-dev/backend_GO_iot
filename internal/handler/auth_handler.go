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
	println("[AUTH] Validating face with Python service...")
	valid, err := h.authSvc.ValidateFaceWithPython(req.FaceImage)
	if err != nil {
		println("[AUTH] Face validation ERROR:", err.Error())
		c.JSON(400, gin.H{"success": false, "error": "Face validation failed: " + err.Error()})
		return
	}
	if !valid {
		println("[AUTH] Face validation FAILED: invalid face")
		c.JSON(400, gin.H{"success": false, "error": "Face validation failed: invalid face detected"})
		return
	}
	println("[AUTH] Face validation SUCCESS")

	// STEP 2: Create user (face sudah valid)
	println("[AUTH] Creating user in database...")
	user, err := h.userSvc.Register(req)
	if err != nil {
		println("[AUTH] User creation ERROR:", err.Error())
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}
	println("[AUTH] User created with ID:", user.UserID)

	// STEP 3: Enroll face & dapat path pkl
	println("[AUTH] Enrolling face to Python service...")
	filename, err := h.authSvc.EnrollFaceWithPython(user.UserID, user.Name, req.FaceImage)
	if err != nil {
		println("[AUTH] Face enrollment ERROR:", err.Error())
		println("[AUTH] Rolling back user creation...")
		h.userSvc.Delete(user.UserID) // Rollback
		c.JSON(500, gin.H{"success": false, "error": "Face enrollment failed: " + err.Error()})
		return
	}
	println("[AUTH] Face enrolled successfully. File:", filename)

	// STEP 4: Update user dengan face_encoding_path
	println("[AUTH] Updating user with face_encoding_path...")
	println("[AUTH] Setting FaceEncodingPath to:", filename)

	err = h.userSvc.UpdateFacePath(user.UserID, filename)
	if err != nil {
		println("[AUTH] Update ERROR:", err.Error())
		c.JSON(500, gin.H{"success": false, "error": "Failed to update user with face path: " + err.Error()})
		return
	}
	user.FaceEncodingPath = filename // Update local object untuk response
	println("[AUTH] User updated successfully!")
	println("[AUTH] Registration completed successfully!")

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
