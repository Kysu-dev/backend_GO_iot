package service

import (
	"log"
	"smarthome-backend/database/models"
	"smarthome-backend/internal/repository"
	"time"
)

type CurtainService interface {
	ProcessCurtain(status string, mode string) error
	GetLatest() (*models.CurtainStatus, error)
}

type curtainService struct {
	repo repository.CurtainRepository
}

func NewCurtainService(r repository.CurtainRepository) CurtainService {
	return &curtainService{repo: r}
}

func (s *curtainService) ProcessCurtain(status string, mode string) error {
	if mode == "" {
		mode = "manual"
	}

	curtain := &models.CurtainStatus{
		Status:    status,
		Mode:      mode,
		Timestamp: time.Now(),
	}

	// Smart CREATE vs UPDATE
	existing, err := s.repo.GetLatest()
	if err != nil {
		// No existing data, create new record
		if err := s.repo.SaveStatus(curtain); err != nil {
			log.Printf("❌ Error creating curtain status: %v", err)
			return err
		}
		log.Printf("✅ Curtain created: %s (Mode: %s)", status, mode)
		return nil
	}

	// Data exists, update it
	if err := s.repo.Update(curtain); err != nil {
		log.Printf("❌ Error updating curtain status: %v", err)
		return err
	}

	log.Printf("✅ Curtain updated: %s (Mode: %s) [prev: %s (%s)]", status, mode, existing.Status, existing.Mode)
	return nil
}

func (s *curtainService) GetLatest() (*models.CurtainStatus, error) {
	return s.repo.GetLatest()
}
