package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"smarthome-backend/database/models"
	"smarthome-backend/internal/service"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
)

// Python service configuration
const (
	PYTHON_SERVICE_URL = "http://localhost:5000"
	UPLOAD_DIR         = "./uploads/camera_captures"
)

// FaceRecognitionResponse from Python service
type FaceRecognitionResponse struct {
	Success    bool    `json:"success"`
	Recognized bool    `json:"recognized"`
	UserID     uint    `json:"user_id"`
	Name       string  `json:"name"`
	Confidence float64 `json:"confidence"`
	Message    string  `json:"message"`
	Error      string  `json:"error,omitempty"`
}

// FaceRecognizeRequest from ESP32-CAM
type FaceRecognizeRequest struct {
	Image string `json:"image" binding:"required"` // base64 encoded image
}

// FaceEnrollRequest for enrolling new face
type FaceEnrollRequest struct {
	UserID uint   `json:"user_id" binding:"required"`
	Name   string `json:"name" binding:"required"`
	Image  string `json:"image" binding:"required"` // base64 encoded image
}

type FaceHandler struct {
	accessLogService service.AccessLogService
	mqttClient       mqtt.Client
}

func NewFaceHandler(accessLogSvc service.AccessLogService, mqttClient mqtt.Client) *FaceHandler {
	// Create upload directory if not exists
	os.MkdirAll(UPLOAD_DIR, os.ModePerm)

	return &FaceHandler{
		accessLogService: accessLogSvc,
		mqttClient:       mqttClient,
	}
}

// RecognizeFace handles face recognition request from ESP32-CAM
// POST /api/face/recognize
func (h *FaceHandler) RecognizeFace(c *gin.Context) {
	var req FaceRecognizeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	log.Println("📸 Face recognition request received")

	// 1. Forward to Python service for recognition
	pythonResp, err := h.sendToPythonService("/recognize", map[string]interface{}{
		"image": req.Image,
	})

	if err != nil {
		log.Printf("❌ Error calling Python service: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Face recognition service unavailable: " + err.Error(),
		})
		return
	}

	// 2. Save to access log
	var userID *uint
	var accessStatus string
	var accessMethod string = "remote" // Face recognition dianggap remote access

	if pythonResp.Recognized {
		userID = &pythonResp.UserID
		accessStatus = "success"
		log.Printf("✅ Face recognized: %s (user_id: %d, confidence: %.2f)",
			pythonResp.Name, pythonResp.UserID, pythonResp.Confidence)
	} else {
		accessStatus = "failed"
		log.Printf("❌ Face not recognized")
	}

	// Save access log
	accessLogReq := models.AccessLogRequest{
		UserID: userID,
		Method: accessMethod,
		Status: accessStatus,
	}

	if err := h.accessLogService.LogAccess(accessLogReq); err != nil {
		log.Printf("⚠️  Failed to save access log: %v", err)
	}

	// 3. If recognized, unlock door via MQTT
	if pythonResp.Recognized {
		unlockPayload := map[string]string{
			"command": "unlock",
			"user_id": fmt.Sprintf("%d", pythonResp.UserID),
			"method":  "face_recognition",
		}

		payload, _ := json.Marshal(unlockPayload)

		token := h.mqttClient.Publish("home/door/control", 0, false, payload)
		token.Wait()

		if token.Error() != nil {
			log.Printf("⚠️  Failed to publish MQTT unlock command: %v", token.Error())
		} else {
			log.Printf("🔓 Door unlock command published via MQTT")
		}
	}

	// 4. Return response
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"recognized": pythonResp.Recognized,
		"user_id":    pythonResp.UserID,
		"name":       pythonResp.Name,
		"confidence": pythonResp.Confidence,
		"message":    pythonResp.Message,
	})
}

// EnrollFace handles face enrollment request
// POST /api/face/enroll
func (h *FaceHandler) EnrollFace(c *gin.Context) {
	var req FaceEnrollRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	log.Printf("📝 Face enrollment request for user_id: %d, name: %s", req.UserID, req.Name)

	// Forward to Python service for enrollment
	pythonResp, err := h.sendToPythonService("/enroll", map[string]interface{}{
		"user_id": req.UserID,
		"name":    req.Name,
		"image":   req.Image,
	})

	if err != nil {
		log.Printf("❌ Error calling Python service: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Face enrollment service unavailable: " + err.Error(),
		})
		return
	}

	if !pythonResp.Success {
		log.Printf("❌ Face enrollment failed: %s", pythonResp.Error)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   pythonResp.Error,
		})
		return
	}

	log.Printf("✅ Face enrolled successfully: %s (user_id: %d)", req.Name, req.UserID)

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: pythonResp.Message,
		Data: gin.H{
			"user_id": req.UserID,
			"name":    req.Name,
		},
	})
}

// ReloadFaces reloads all known faces in Python service
// POST /api/face/reload
func (h *FaceHandler) ReloadFaces(c *gin.Context) {
	log.Println("🔄 Reloading faces in Python service...")

	pythonResp, err := h.sendToPythonService("/reload", nil)

	if err != nil {
		log.Printf("❌ Error calling Python service: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Face service unavailable: " + err.Error(),
		})
		return
	}

	log.Printf("✅ Faces reloaded successfully")

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: pythonResp.Message,
	})
}

// GetAccessLogs returns recent access logs
// GET /api/face/logs
func (h *FaceHandler) GetAccessLogs(c *gin.Context) {
	// This will be handled by access_log_handler
	// Redirect or call the service directly
	logs, err := h.accessLogService.GetAll(100) // Get last 100 logs

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Access logs retrieved",
		Data:    logs,
	})
}

// sendToPythonService sends HTTP request to Python service
func (h *FaceHandler) sendToPythonService(endpoint string, data map[string]interface{}) (*FaceRecognitionResponse, error) {
	url := PYTHON_SERVICE_URL + endpoint

	var req *http.Request
	var err error

	if data != nil {
		jsonData, _ := json.Marshal(data)
		req, err = http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest("POST", url, nil)
		if err != nil {
			return nil, err
		}
	}

	// Set timeout 30 seconds (face recognition bisa lama)
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Python service: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var result FaceRecognitionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &result, nil
}

// SaveBase64Image saves base64 image to file (helper function)
func (h *FaceHandler) SaveBase64Image(base64Image string) (string, error) {
	// This is optional - for debugging/logging purposes
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("capture_%s.jpg", timestamp)
	filepath := filepath.Join(UPLOAD_DIR, filename)

	// Decode base64
	// ... implementation if needed

	return filepath, nil
}
