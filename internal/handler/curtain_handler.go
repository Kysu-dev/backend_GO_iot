package handler

import (
	"smarthome-backend/database/models"
	"smarthome-backend/internal/service"

	// "strconv" -> Hapus ini karena tidak dipakai lagi

	"github.com/gin-gonic/gin"
)

type CurtainHandler struct {
	svc service.CurtainService
}

func NewCurtainHandler(s service.CurtainService) *CurtainHandler {
	return &CurtainHandler{svc: s}
}

func (h *CurtainHandler) Create(c *gin.Context) {
	var req models.CurtainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	err := h.svc.ProcessCurtain(req.Status, req.Mode)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "Failed to save curtain status"})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "Curtain status saved"})
}

func (h *CurtainHandler) GetLatest(c *gin.Context) {
	data, err := h.svc.GetLatest()
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "No data found"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": data})
}

// --- FUNGSI GetAll DIHAPUS ---
// Karena kita hanya menyimpan 1 baris data (Upsert),
// maka tidak ada history yang bisa diambil.
