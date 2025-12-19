package service

import (
	"errors"
	"fmt"
	"net/http"
	"smarthome-backend/database/models"
	"smarthome-backend/internal/repository"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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
	CreateAdmin(req models.UserRequest) (*models.User, error)
	GetAdmins() ([]models.User, error)
	GetAdminByID(id uint) (*models.User, error)
	UpdateAdmin(id uint, req models.UpdateUserRequest) (*models.User, error)
	DeleteAdmin(id uint) error
}

type userService struct {
	repo repository.UserRepository
}

// normalizeRole maps external labels to DB enum values.
// Accepts "user" as alias for "member" to fit the enum.
func normalizeRole(role string) string {
	switch role {
	case "user":
		return "member"
	case "member", "admin":
		return role
	default:
		return role
	}
}

func normalizeStatus(status string) (string, error) {
	switch status {
	case "pending", "active", "suspended":
		return status, nil
	case "rejected":
		return "", errors.New("status 'rejected' is not supported; use pending/active/suspended")
	default:
		return "", errors.New("invalid status")
	}
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

	role := normalizeRole(req.Role)
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

// Admin CRUD helpers
func (s *userService) CreateAdmin(req models.UserRequest) (*models.User, error) {
	// Reuse validation: email uniqueness
	if existing, _ := s.repo.FindByEmail(req.Email); existing != nil && existing.UserID > 0 {
		return nil, errors.New("email already registered")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	admin := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     "admin",
		Status:   "active",
	}

	if err := s.repo.Create(admin); err != nil {
		return nil, err
	}

	return admin, nil
}

func (s *userService) GetAdmins() ([]models.User, error) {
	all, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	admins := make([]models.User, 0)
	for _, u := range all {
		if u.Role == "admin" {
			admins = append(admins, u)
		}
	}
	return admins, nil
}

func (s *userService) GetAdminByID(id uint) (*models.User, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if user.Role != "admin" {
		return nil, errors.New("user is not admin")
	}
	return user, nil
}

func (s *userService) UpdateAdmin(id uint, req models.UpdateUserRequest) (*models.User, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}
	if user.Role != "admin" {
		return nil, errors.New("user is not admin")
	}

	// If email changes, ensure unique
	if user.Email != req.Email {
		existing, err := s.repo.FindByEmail(req.Email)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}
		} else if existing != nil && existing.UserID != id {
			return nil, errors.New("email already in use")
		}
	}

	normalizedStatus, err := normalizeStatus(req.Status)
	if err != nil {
		return nil, err
	}

	user.Name = req.Name
	user.Email = req.Email
	user.Role = "admin" // force admin
	user.Status = normalizedStatus

	if err := s.repo.Update(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) DeleteAdmin(id uint) error {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if user.Role != "admin" {
		return errors.New("user is not admin")
	}
	return s.repo.Delete(id)
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
		existing, err := s.repo.FindByEmail(req.Email)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}
		} else if existing != nil && existing.UserID != id {
			return nil, errors.New("email already in use")
		}
	}

	normalizedRole := normalizeRole(req.Role)
	if normalizedRole != "admin" && normalizedRole != "member" {
		return nil, errors.New("invalid role")
	}

	normalizedStatus, err := normalizeStatus(req.Status)
	if err != nil {
		return nil, err
	}

	// Update fields (face_encoding_path is NOT touched)
	user.Name = req.Name
	user.Email = req.Email
	user.Role = normalizedRole
	user.Status = normalizedStatus

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
