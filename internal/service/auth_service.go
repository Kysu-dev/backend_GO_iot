package service

import (
	"errors"
	"fmt"
	"smarthome-backend/database/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService interface {
	GenerateToken(user *models.User) (string, error)
	ValidateToken(token string) (*models.User, error)
	EnrollFaceWithPython(userID uint, name string, faceImage string) (string, error)
	ValidateFaceWithPython(faceImage string) (bool, error)
}

type authService struct {
	pythonClient *PythonFaceClient
	jwtSecret    string
}

func NewAuthService(pythonURL string, secret string) AuthService {
	return &authService{
		pythonClient: NewPythonFaceClient(pythonURL),
		jwtSecret:    secret,
	}
}

func (s *authService) GenerateToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.UserID,
		"email":   user.Email,
		"name":    user.Name,
		"role":    user.Role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *authService) ValidateToken(tokenString string) (*models.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		user := &models.User{
			UserID: uint(claims["user_id"].(float64)),
			Email:  claims["email"].(string),
			Name:   claims["name"].(string),
			Role:   claims["role"].(string),
		}
		return user, nil
	}

	return nil, errors.New("invalid token")
}

func (s *authService) EnrollFaceWithPython(userID uint, name string, faceImage string) (string, error) {
	resp, err := s.pythonClient.EnrollFace(int(userID), name, faceImage)
	if err != nil {
		return "", fmt.Errorf("failed to enroll face: %w", err)
	}

	if !resp.Success {
		return "", errors.New(resp.Message)
	}

	return resp.File, nil
}

func (s *authService) ValidateFaceWithPython(faceImage string) (bool, error) {
	resp, err := s.pythonClient.ValidateFace(faceImage)
	if err != nil {
		return false, fmt.Errorf("failed to validate face: %w", err)
	}

	if !resp.Valid {
		return false, errors.New(resp.Message)
	}

	return true, nil
}
