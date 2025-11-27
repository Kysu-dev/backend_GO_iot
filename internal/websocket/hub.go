package websocket

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin:     func(r *http.Request) bool { return true },
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mutex      sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

// Goroutine: Standby kirim data ke frontend
func (h *Hub) Run() {
	log.Println("🚀 WebSocket Hub started")
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			log.Printf("✅ Client connected. Total clients: %d", len(h.clients))

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
				log.Printf("🔴 Client disconnected. Total clients: %d", len(h.clients))
			}
			h.mutex.Unlock()

		case msg := <-h.broadcast:
			h.mutex.Lock()
			for client := range h.clients {
				err := client.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					log.Printf("❌ Error sending to client: %v", err)
					client.Close()
					delete(h.clients, client)
				}
			}
			h.mutex.Unlock()
		}
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
		log.Println("❌ WebSocket upgrade error:", err)
		return
	}

	// Register client
	h.register <- conn

	// Start read pump (keep connection alive)
	go h.readPump(conn)

	log.Printf("✅ New WebSocket client connected from %s", c.ClientIP())
}

// Read pump untuk handle disconnect
func (h *Hub) readPump(conn *websocket.Conn) {
	defer func() {
		h.unregister <- conn
	}()

	// Set read deadline
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("⚠️ Unexpected close: %v", err)
			}
			break
		}
	}
}

// Ping timer untuk keep-alive
func (h *Hub) StartPingTimer() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		h.mutex.Lock()
		for client := range h.clients {
			err := client.WriteMessage(websocket.PingMessage, []byte{})
			if err != nil {
				log.Printf("❌ Ping failed: %v", err)
				client.Close()
				delete(h.clients, client)
			}
		}
		h.mutex.Unlock()
	}
}
