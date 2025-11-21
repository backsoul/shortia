package api

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for local development
	},
}

type WSMessage struct {
	Type    string      `json:"type"`
	Status  string      `json:"status,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
}

// Video-specific WebSocket connections
type VideoWSManager struct {
	clients map[string]map[*websocket.Conn]bool // videoID -> connections
	mu      sync.RWMutex
}

var wsManager = &VideoWSManager{
	clients: make(map[string]map[*websocket.Conn]bool),
}

func VideoWebSocketHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		videoID := c.Param("id")

		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("Failed to upgrade to WebSocket: %v", err)
			return
		}
		defer ws.Close()

		// Register client for this video
		wsManager.mu.Lock()
		if wsManager.clients[videoID] == nil {
			wsManager.clients[videoID] = make(map[*websocket.Conn]bool)
		}
		wsManager.clients[videoID][ws] = true
		wsManager.mu.Unlock()

		log.Printf("Client connected to video %s. Total clients: %d", videoID, len(wsManager.clients[videoID]))

		// Keep connection alive and handle messages
		for {
			_, _, err := ws.ReadMessage()
			if err != nil {
				log.Printf("WebSocket error: %v", err)
				wsManager.mu.Lock()
				delete(wsManager.clients[videoID], ws)
				if len(wsManager.clients[videoID]) == 0 {
					delete(wsManager.clients, videoID)
				}
				wsManager.mu.Unlock()
				break
			}
		}
	}
}

// BroadcastVideoStatus sends status updates to all clients watching a specific video
func BroadcastVideoStatus(videoID string, status string) {
	wsManager.mu.RLock()
	clients := wsManager.clients[videoID]
	wsManager.mu.RUnlock()

	if clients == nil {
		return
	}

	msg := WSMessage{
		Type:   "status",
		Status: status,
	}

	wsManager.mu.Lock()
	defer wsManager.mu.Unlock()

	for client := range clients {
		err := client.WriteJSON(msg)
		if err != nil {
			log.Printf("WebSocket write error: %v", err)
			client.Close()
			delete(clients, client)
		}
	}

	if len(clients) == 0 {
		delete(wsManager.clients, videoID)
	}
}

// BroadcastProgress sends progress updates to all clients watching a specific video
func BroadcastProgress(videoID string, status string, progress int, message string) {
	BroadcastVideoStatus(videoID, status)
}
