package service

import (
	"errors"
	"log"
	"smarthome-backend/database/models"
	"smarthome-backend/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(req models.UserRequest) (*models.User, error)
	Login(req models.LoginRequest) (*models.User, error)
	GetByID(id uint) (*models.User, error)
	GetAll() ([]models.User, error)
	Update(user *models.User) error
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
	}

	err = s.repo.Create(user)
	if err != nil {
		log.Printf("❌ Error creating user: %v", err)
		return nil, err
	}

	log.Printf("✅ User registered: %s (%s)", user.Email, user.Role)
	return user, nil
}

func (s *userService) Login(req models.LoginRequest) (*models.User, error) {
	user, err := s.repo.FindByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	log.Printf("✅ User logged in: %s", user.Email)
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

func (s *userService) Delete(id uint) error {
	return s.repo.Delete(id)
}
