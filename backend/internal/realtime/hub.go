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

func (h *Hub) Add(userID int64, conn *websocket.Conn) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[userID] == nil {
		h.clients[userID] = make(map[*websocket.Conn]struct{})
	}

	wasOffline := len(h.clients[userID]) == 0

	h.clients[userID][conn] = struct{}{}

	return wasOffline
}

func (h *Hub) Remove(userID int64, conn *websocket.Conn) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[userID] == nil {
		return false
	}

	delete(h.clients[userID], conn)

	if len(h.clients[userID]) == 0 {
		delete(h.clients, userID)
		return true
	}

	return false
}

func (h *Hub) IsOnline(userID int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.clients[userID]) > 0
}

func (h *Hub) OnlineUserIDs() []int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	ids := make([]int64, 0, len(h.clients))

	for userID := range h.clients {
		ids = append(ids, userID)
	}

	return ids
}

func (h *Hub) SendToUser(userID int64, event Event) {
	payload, err := json.Marshal(event)
	if err != nil {
		return
	}

	h.writeMu.Lock()
	defer h.writeMu.Unlock()

	h.mu.RLock()
	defer h.mu.RUnlock()

	for conn := range h.clients[userID] {
		_ = conn.WriteMessage(websocket.TextMessage, payload)
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

func (h *Hub) Broadcast(event Event) {
	payload, err := json.Marshal(event)
	if err != nil {
		return
	}

	h.writeMu.Lock()
	defer h.writeMu.Unlock()

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, connections := range h.clients {
		for conn := range connections {
			_ = conn.WriteMessage(websocket.TextMessage, payload)
		}
	}
}
