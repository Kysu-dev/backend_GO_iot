package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// PythonFaceClient handles communication with Python Face Recognition Service
type PythonFaceClient struct {
	baseURL string
	client  *http.Client
}

// NewPythonFaceClient creates a new PythonFaceClient instance
func NewPythonFaceClient(baseURL string) *PythonFaceClient {
	return &PythonFaceClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ValidateFaceRequest represents the request payload for face validation
type ValidateFaceRequest struct {
	Image string `json:"image"` // Base64 encoded image with data URI scheme
}

// ValidateFaceResponse represents the response from face validation
type ValidateFaceResponse struct {
	Valid         bool   `json:"valid"`
	FacesDetected int    `json:"faces_detected"`
	Message       string `json:"message"`
}

// ValidateFace validates if the image contains exactly one face
func (c *PythonFaceClient) ValidateFace(imageBase64 string) (*ValidateFaceResponse, error) {
	reqBody := ValidateFaceRequest{
		Image: imageBase64,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.client.Post(
		fmt.Sprintf("%s/validate-face", c.baseURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to call Python service: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Python service error (status %d): %s", resp.StatusCode, string(body))
	}

	var result ValidateFaceResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// EnrollFaceRequest represents the request payload for face enrollment
type EnrollFaceRequest struct {
	UserID int    `json:"user_id"`
	Name   string `json:"name"`
	Image  string `json:"image"` // Base64 encoded image with data URI scheme
}

// EnrollFaceResponse represents the response from face enrollment
type EnrollFaceResponse struct {
	Success bool   `json:"success"`
	File    string `json:"file"`
	Message string `json:"message"`
}

// EnrollFace enrolls a new face for the given user
func (c *PythonFaceClient) EnrollFace(userID int, name, imageBase64 string) (*EnrollFaceResponse, error) {
	reqBody := EnrollFaceRequest{
		UserID: userID,
		Name:   name,
		Image:  imageBase64,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.client.Post(
		fmt.Sprintf("%s/enroll-base64", c.baseURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to call Python service: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Python service error (status %d): %s", resp.StatusCode, string(body))
	}

	var result EnrollFaceResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// RecognizeFaceRequest represents the request payload for face recognition
type RecognizeFaceRequest struct {
	Image string `json:"image"` // Base64 encoded image with data URI scheme
}

// RecognizeFaceResponse represents the response from face recognition
type RecognizeFaceResponse struct {
	Recognized bool    `json:"recognized"`
	UserID     int     `json:"user_id,omitempty"`
	Name       string  `json:"name,omitempty"`
	Confidence float64 `json:"confidence,omitempty"`
	Message    string  `json:"message"`
}

// RecognizeFace recognizes a face from the given image
func (c *PythonFaceClient) RecognizeFace(imageBase64 string) (*RecognizeFaceResponse, error) {
	reqBody := RecognizeFaceRequest{
		Image: imageBase64,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.client.Post(
		fmt.Sprintf("%s/recognize-base64", c.baseURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to call Python service: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Python service error (status %d): %s", resp.StatusCode, string(body))
	}

	var result RecognizeFaceResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// EnrolledFace represents a single enrolled face
type EnrolledFace struct {
	File   string `json:"file"`
	UserID int    `json:"user_id"`
	Name   string `json:"name"`
}

// GetEnrolledFacesResponse represents the response from getting enrolled faces
type GetEnrolledFacesResponse struct {
	Faces []EnrolledFace `json:"faces"`
	Total int            `json:"total"`
}

// GetEnrolledFaces retrieves all enrolled faces from the Python service
func (c *PythonFaceClient) GetEnrolledFaces() (*GetEnrolledFacesResponse, error) {
	resp, err := c.client.Get(fmt.Sprintf("%s/faces", c.baseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to call Python service: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Python service error (status %d): %s", resp.StatusCode, string(body))
	}

	var result GetEnrolledFacesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// ReloadFaces triggers the Python service to reload face encodings from disk
func (c *PythonFaceClient) ReloadFaces() error {
	resp, err := c.client.Post(
		fmt.Sprintf("%s/reload", c.baseURL),
		"application/json",
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to call Python service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Python service error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// HealthCheck checks if the Python service is running
func (c *PythonFaceClient) HealthCheck() error {
	resp, err := c.client.Get(fmt.Sprintf("%s/health", c.baseURL))
	if err != nil {
		return fmt.Errorf("Python service unavailable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Python service unhealthy (status %d)", resp.StatusCode)
	}

	return nil
}
