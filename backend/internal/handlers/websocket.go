package handlers

import (
	"net/http"

	"github.com/gorilla/websocket"

	"go-chat-app/backend/internal/middleware"
	"go-chat-app/backend/internal/realtime"
)

type WebSocketHandler struct {
	Auth           middleware.Auth
	Hub            *realtime.Hub
	FrontendOrigin string
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

	wasOffline := h.Hub.Add(userID, conn)

	defer func() {
		wentOffline := h.Hub.Remove(userID, conn)

		_ = conn.Close()

		if wentOffline {
			h.Hub.Broadcast(realtime.Event{
				Type: "presence",
				Data: map[string]any{
					"user_id": userID,
					"online":  false,
				},
			})
		}
	}()

	// Send the currently online users to the newly connected user.
	h.Hub.SendToUser(userID, realtime.Event{
		Type: "online_users",
		Data: h.Hub.OnlineUserIDs(),
	})

	// Notify other connected users when this user becomes online.
	if wasOffline {
		h.Hub.Broadcast(realtime.Event{
			Type: "presence",
			Data: map[string]any{
				"user_id": userID,
				"online":  true,
			},
		})
	}

	// Keep reading so we detect browser disconnects.
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}
