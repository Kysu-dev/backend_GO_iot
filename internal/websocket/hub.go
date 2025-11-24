package websocket

import (
	"log"
	"net/http"
	"sync"
	"github.com/gorilla/websocket"
	"github.com/gin-gonic/gin"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Hub struct {
	clients   map[*websocket.Conn]bool
	broadcast chan []byte
	mutex     sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan []byte),
	}
}

// Goroutine: Standby kirim data ke frontend
func (h *Hub) Run() {
	for {
		msg := <-h.broadcast
		h.mutex.Lock()
		for client := range h.clients {
			client.WriteMessage(websocket.TextMessage, msg)
		}
		h.mutex.Unlock()
	}
}

// Fungsi Penyebar Data
func (h *Hub) BroadcastData(data []byte) {
	h.broadcast <- data
}

// Endpoint WebSocket
func (h *Hub) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket Error:", err)
		return
	}
	h.mutex.Lock()
	h.clients[conn] = true
	h.mutex.Unlock()
	log.Println("Client WebSocket Connected!")
}