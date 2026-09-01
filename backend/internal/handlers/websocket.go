package handlers

import (
	"net/http"

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

	// Keep reading so we detect browser disconnects.
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}