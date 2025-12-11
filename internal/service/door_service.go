package service

import (
	"log"
	"smarthome-backend/database/models"
	"smarthome-backend/internal/repository"
)

type DoorService interface {
	ProcessDoor(status, method string, userID *uint) error
	GetLatest() (*models.DoorStatus, error)
	GetHistory(limit int) ([]models.DoorStatus, error)
}

type doorService struct {
	repo          repository.DoorRepository
	accessLogRepo repository.AccessLogRepository
}

func NewDoorService(r repository.DoorRepository, accessLogRepo repository.AccessLogRepository) DoorService {
	return &doorService{repo: r, accessLogRepo: accessLogRepo}
}

func (s *doorService) ProcessDoor(status, method string, userID *uint) error {
	door := &models.DoorStatus{
		Status: status,
		Method: method,
	}

	// Cek apakah sudah ada data
	existing, err := s.repo.GetLatest()

	// Jika tidak ada data atau error "record not found", maka CREATE
	if err != nil {
		// Belum ada data, CREATE
		err = s.repo.Create(door)
		if err != nil {
			log.Printf("❌ Error creating door status: %v", err)
			return err
		}
		log.Printf("✅ Door status created: %s via %s", status, method)
		return nil
	}

	// Jika data ada dan DoorID > 0, maka UPDATE
	if existing != nil && existing.DoorID > 0 {
		err = s.repo.Update(door)
		if err != nil {
			log.Printf("❌ Error updating door status: %v", err)
			return err
		}
		log.Printf("✅ Door status updated: %s via %s", status, method)

		// ⭐ Save to access log history (async) dengan user_id
		go s.saveAccessLog(status, method, userID)

		return nil
	}

	// Fallback: CREATE jika kondisi tidak terpenuhi
	err = s.repo.Create(door)
	if err != nil {
		log.Printf("❌ Error creating door status: %v", err)
		return err
	}
	log.Printf("✅ Door status created: %s via %s", status, method)
	return nil
}

func (s *doorService) GetLatest() (*models.DoorStatus, error) {
	return s.repo.GetLatest()
}

func (s *doorService) GetHistory(limit int) ([]models.DoorStatus, error) {
	return s.repo.GetHistory(limit)
}

// saveAccessLog - Save door access to history
func (s *doorService) saveAccessLog(status, method string, userID *uint) {
	// Avoid duplicate access logs for face recognition; handler already logs those
	if method == "face" {
		return
	}
	// Determine if access was successful (unlocked = success)
	accessStatus := "failed"
	if status == "unlocked" {
		accessStatus = "success"
	}

	accessLog := &models.AccessLog{
		Method:    method,
		Status:    accessStatus,
		UserID:    userID, // ⭐ Simpan user_id
		ImagePath: "",
	}

	err := s.accessLogRepo.Create(accessLog)
	if err != nil {
		log.Printf("⚠️ Failed to save access log: %v", err)
	} else {
		if userID != nil {
			log.Printf("📝 Access log saved: %s (%s) by user_id: %d", method, accessStatus, *userID)
		} else {
			log.Printf("📝 Access log saved: %s (%s)", method, accessStatus)
		}
	}
}
