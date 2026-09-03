package realtime

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

type Event struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type Hub struct {
	mu      sync.RWMutex
	writeMu sync.Mutex
	clients map[int64]map[*websocket.Conn]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[int64]map[*websocket.Conn]struct{}),
	}
}

func (h *Hub) Add(userID int64, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[userID] == nil {
		h.clients[userID] = make(map[*websocket.Conn]struct{})
	}

	h.clients[userID][conn] = struct{}{}
}

func (h *Hub) Remove(userID int64, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[userID] != nil {
		delete(h.clients[userID], conn)

		if len(h.clients[userID]) == 0 {
			delete(h.clients, userID)
		}
	}
}

func (h *Hub) SendToUsers(userIDs []int64, event Event) {
	payload, err := json.Marshal(event)
	if err != nil {
		return
	}

	h.writeMu.Lock()
	defer h.writeMu.Unlock()

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, userID := range userIDs {
		for conn := range h.clients[userID] {
			_ = conn.WriteMessage(websocket.TextMessage, payload)
		}
	}
}

func (h *Hub) Ping(conn *websocket.Conn) error {
	h.writeMu.Lock()
	defer h.writeMu.Unlock()

	return conn.WriteMessage(websocket.PingMessage, nil)
}
