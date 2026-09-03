package handlers

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"go-chat-app/backend/internal/middleware"
	"go-chat-app/backend/internal/realtime"
)

type WebSocketHandler struct {
	Auth            middleware.Auth
	Hub             *realtime.Hub
	FrontendOrigin  string
}

func (h WebSocketHandler) Connect(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	userID, err := h.Auth.ParseUserID(token)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return r.Header.Get("Origin") == h.FrontendOrigin
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	h.Hub.Add(userID, conn)

	defer func() {
		h.Hub.Remove(userID, conn)
		_ = conn.Close()
	}()

	const (
		pongWait   = 60 * time.Second
		pingPeriod = 30 * time.Second
	)

	conn.SetReadDeadline(time.Now().Add(pongWait))

	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				if err := h.Hub.Ping(conn); err != nil {
					close(done)
					return
				}

			case <-done:
				return
			}
		}
	}()

	// Keep reading so we detect browser disconnects.
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}