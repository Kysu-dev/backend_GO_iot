package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"smarthome-backend/database/models"
)

type AuthService interface {
	GenerateToken(user *models.User) (string, error)
	ValidateToken(token string) (*models.User, error)
	EnrollFaceWithPython(userID uint, name string, faceImage string) (string, error)
	ValidateFaceWithPython(faceImage string) (bool, error)
}

type authService struct {
	pythonServiceURL string
	jwtSecret        string
}

func NewAuthService(pythonURL string, secret string) AuthService {
	return &authService{
		pythonServiceURL: pythonURL,
		jwtSecret:        secret,
	}
}

func (s *authService) GenerateToken(user *models.User) (string, error) {
	return fmt.Sprintf("jwt_token_%d", user.UserID), nil
}

func (s *authService) ValidateToken(token string) (*models.User, error) {
	return nil, errors.New("not implemented")
}

func (s *authService) EnrollFaceWithPython(userID uint, name string, faceImage string) (string, error) {
	url := fmt.Sprintf("%s/enroll-base64", s.pythonServiceURL)

	payload := map[string]interface{}{
		"user_id": userID,
		"name":    name,
		"image":   faceImage,
	}

	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if resp.StatusCode != 200 {
		return "", errors.New(result["error"].(string))
	}

	filename := fmt.Sprintf("%d_%s_1.pkl", userID, name)
	return filename, nil
}

func (s *authService) ValidateFaceWithPython(faceImage string) (bool, error) {
	url := fmt.Sprintf("%s/validate-face", s.pythonServiceURL)

	payload := map[string]interface{}{
		"image": faceImage,
	}

	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if resp.StatusCode != 200 {
		return false, errors.New(result["error"].(string))
	}

	return result["face_valid"].(bool), nil
}
