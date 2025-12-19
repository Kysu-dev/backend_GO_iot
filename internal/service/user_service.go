package service

import (
	"errors"
	"fmt"
	"net/http"
	"smarthome-backend/database/models"
	"smarthome-backend/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(req models.UserRequest) (*models.User, error)
	Login(req models.LoginRequest) (*models.User, error)
	GetByID(id uint) (*models.User, error)
	GetAll() ([]models.User, error)
	GetPending() ([]models.User, error)
	Approve(id uint) error
	Reject(id uint) error
	Update(user *models.User) error
	UpdateUser(id uint, req models.UpdateUserRequest) (*models.User, error)
	UpdateFacePath(userID uint, facePath string) error
	Delete(id uint) error
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(r repository.UserRepository) UserService {
	return &userService{repo: r}
}

func (s *userService) Register(req models.UserRequest) (*models.User, error) {
	// Check if email already exists
	existing, _ := s.repo.FindByEmail(req.Email)
	if existing != nil && existing.UserID > 0 {
		return nil, errors.New("email already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Set default role if not provided
	role := req.Role
	if role == "" {
		role = "member"
	}

	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     role,
		Status:   "pending",
	}

	err = s.repo.Create(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) Login(req models.LoginRequest) (*models.User, error) {
	user, err := s.repo.FindByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if user.Status != "active" {
		return nil, errors.New("account not active")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	return user, nil
}

func (s *userService) GetByID(id uint) (*models.User, error) {
	return s.repo.FindByID(id)
}

func (s *userService) GetAll() ([]models.User, error) {
	return s.repo.GetAll()
}

func (s *userService) Update(user *models.User) error {
	return s.repo.Update(user)
}

func (s *userService) UpdateFacePath(userID uint, facePath string) error {
	return s.repo.UpdateFacePath(userID, facePath)
}

func (s *userService) UpdateUser(id uint, req models.UpdateUserRequest) (*models.User, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Check if email is being changed and already exists
	if user.Email != req.Email {
		existing, _ := s.repo.FindByEmail(req.Email)
		if existing != nil && existing.UserID != id {
			return nil, errors.New("email already in use")
		}
	}

	// Update fields (face_encoding_path is NOT touched)
	user.Name = req.Name
	user.Email = req.Email
	user.Role = req.Role
	user.Status = req.Status

	err = s.repo.Update(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) Delete(id uint) error {
	// Get user first to check if they have face encoding
	user, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	// If user has face encoding, delete it from Python service first
	if user.FaceEncodingPath != "" {
		pythonServiceURL := "http://localhost:5000"
		deleteURL := fmt.Sprintf("%s/faces/%d", pythonServiceURL, id)

		req, err := http.NewRequest("DELETE", deleteURL, nil)
		if err != nil {
			// Log error but continue with user deletion
			fmt.Printf("⚠️  Failed to create request to delete face encoding: %v\n", err)
		} else {
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				// Log error but continue with user deletion
				fmt.Printf("⚠️  Failed to delete face encoding from Python service: %v\n", err)
			} else {
				defer resp.Body.Close()
				if resp.StatusCode == 200 {
					fmt.Printf("✅ Face encoding deleted for user_id: %d\n", id)
				} else {
					fmt.Printf("⚠️  Python service returned status %d when deleting face\n", resp.StatusCode)
				}
			}
		}
	}

	// Delete user from database
	return s.repo.Delete(id)
}

func (s *userService) GetPending() ([]models.User, error) {
	allUsers, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	pending := []models.User{}
	for _, user := range allUsers {
		if user.Status == "pending" {
			pending = append(pending, user)
		}
	}
	return pending, nil
}

func (s *userService) Approve(id uint) error {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	user.Status = "active"
	return s.repo.Update(user)
}

func (s *userService) Reject(id uint) error {
	return s.repo.Delete(id)
}
