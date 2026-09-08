package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"

	"go-chat-app/backend/internal/middleware"
	"go-chat-app/backend/internal/realtime"
)

type WebSocketHandler struct {
	Auth           middleware.Auth
	Hub            *realtime.Hub
	DB             *pgxpool.Pool
	FrontendOrigin string
}

type typingEvent struct {
	Type string `json:"type"`
	Data struct {
		ConversationID int64 `json:"conversation_id"`
	} `json:"data"`
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
			h.Hub.Broadcast(
				realtime.Event{
					Type: "presence",
					Data: map[string]any{
						"user_id": userID,
						"online":  false,
					},
				},
			)
		}
	}()

	onlineUsers := h.Hub.OnlineUserIDs()

	h.Hub.SendToUser(
		userID,
		realtime.Event{
			Type: "online_users",
			Data: onlineUsers,
		},
	)

	if wasOffline {
		h.Hub.Broadcast(
			realtime.Event{
				Type: "presence",
				Data: map[string]any{
					"user_id": userID,
					"online":  true,
				},
			},
		)
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var event typingEvent

		if err := json.Unmarshal(message, &event); err != nil {
			continue
		}

		if event.Type != "typing_start" && event.Type != "typing_stop" {
			continue
		}

		if event.Data.ConversationID == 0 {
			continue
		}

		participantIDs := h.participantIDs(
			r,
			event.Data.ConversationID,
			userID,
		)

		if len(participantIDs) == 0 {
			continue
		}

		h.Hub.SendToUsers(
			participantIDs,
			realtime.Event{
				Type: event.Type,
				Data: map[string]any{
					"conversation_id": event.Data.ConversationID,
					"user_id":         userID,
				},
			},
		)
	}
}

func (h WebSocketHandler) participantIDs(
	r *http.Request,
	conversationID int64,
	currentUserID int64,
) []int64 {
	rows, err := h.DB.Query(r.Context(), `
		SELECT user_id
		FROM conversation_participants
		WHERE conversation_id = $1
		AND user_id <> $2
	`, conversationID, currentUserID)
	if err != nil {
		return nil
	}

	defer rows.Close()

	ids := []int64{}

	for rows.Next() {
		var id int64

		if rows.Scan(&id) == nil {
			ids = append(ids, id)
		}
	}

	return ids
}