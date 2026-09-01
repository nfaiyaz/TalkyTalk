package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"go-chat-app/backend/internal/middleware"
	"go-chat-app/backend/internal/models"
	"go-chat-app/backend/internal/realtime"
)

type MessageHandler struct {
	DB  *pgxpool.Pool
	Hub *realtime.Hub
}

type sendMessageRequest struct {
	Body string `json:"body"`
}

func (h MessageHandler) List(w http.ResponseWriter, r *http.Request) {
	conversationID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)

	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid conversation id")
		return
	}

	userID := middleware.UserID(r)

	if !h.isParticipant(r, conversationID, userID) {
		writeError(w, http.StatusForbidden, "not a participant")
		return
	}

	rows, err := h.DB.Query(r.Context(), `
SELECT m.id, m.conversation_id, m.sender_id, u.name, m.body, m.created_at
FROM messages m
JOIN users u ON u.id = m.sender_id
WHERE m.conversation_id = $1
ORDER BY m.created_at ASC
LIMIT 200
`, conversationID)

	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load messages")
		return
	}

	defer rows.Close()

	messages := []models.Message{}

	for rows.Next() {
		var m models.Message

		if err := rows.Scan(
			&m.ID,
			&m.ConversationID,
			&m.SenderID,
			&m.SenderName,
			&m.Body,
			&m.CreatedAt,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "could not read messages")
			return
		}

		messages = append(messages, m)
	}

	writeJSON(w, http.StatusOK, messages)
}

func (h MessageHandler) Send(w http.ResponseWriter, r *http.Request) {
	conversationID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)

	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid conversation id")
		return
	}

	userID := middleware.UserID(r)

	if !h.isParticipant(r, conversationID, userID) {
		writeError(w, http.StatusForbidden, "not a participant")
		return
	}

	var input sendMessageRequest

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	input.Body = strings.TrimSpace(input.Body)

	if len(input.Body) == 0 || len(input.Body) > 4000 {
		writeError(w, http.StatusBadRequest, "message must be 1-4000 characters")
		return
	}

	var m models.Message

	err = h.DB.QueryRow(r.Context(), `
INSERT INTO messages (conversation_id, sender_id, body)
VALUES ($1, $2, $3)
RETURNING id, conversation_id, sender_id, body, created_at
`, conversationID, userID, input.Body).Scan(
		&m.ID,
		&m.ConversationID,
		&m.SenderID,
		&m.Body,
		&m.CreatedAt,
	)

	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not save message")
		return
	}

	_ = h.DB.QueryRow(
		r.Context(),
		`SELECT name FROM users WHERE id = $1`,
		userID,
	).Scan(&m.SenderName)

	participantIDs := h.participantIDs(r, conversationID)

	h.Hub.SendToUsers(
		participantIDs,
		realtime.Event{
			Type: "new_message",
			Data: m,
		},
	)

	writeJSON(w, http.StatusCreated, m)
}

func (h MessageHandler) isParticipant(
	r *http.Request,
	conversationID,
	userID int64,
) bool {
	var exists bool

	err := h.DB.QueryRow(r.Context(), `
SELECT EXISTS(
SELECT 1 FROM conversation_participants
WHERE conversation_id = $1 AND user_id = $2
)
`, conversationID, userID).Scan(&exists)

	return err == nil && exists
}

func (h MessageHandler) participantIDs(
	r *http.Request,
	conversationID int64,
) []int64 {
	rows, err := h.DB.Query(r.Context(), `
SELECT user_id FROM conversation_participants WHERE conversation_id = $1
`, conversationID)

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