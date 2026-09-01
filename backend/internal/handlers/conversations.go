package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"go-chat-app/backend/internal/middleware"
	"go-chat-app/backend/internal/models"
)

type ConversationHandler struct {
	DB *pgxpool.Pool
}

type createConversationRequest struct {
	UserID int64 `json:"user_id"`
}

func (h ConversationHandler) Create(w http.ResponseWriter, r *http.Request) {
	currentUserID := middleware.UserID(r)

	var input createConversationRequest

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.UserID == 0 {
		writeError(w, http.StatusBadRequest, "valid user_id is required")
		return
	}

	if input.UserID == currentUserID {
		writeError(w, http.StatusBadRequest, "you cannot chat with yourself")
		return
	}

	var existingID int64

	err := h.DB.QueryRow(r.Context(), `
SELECT cp1.conversation_id
FROM conversation_participants cp1
JOIN conversation_participants cp2
ON cp1.conversation_id = cp2.conversation_id
WHERE cp1.user_id = $1 AND cp2.user_id = $2
AND (SELECT COUNT(*) FROM conversation_participants x
WHERE x.conversation_id = cp1.conversation_id) = 2
LIMIT 1
`, currentUserID, input.UserID).Scan(&existingID)

	if err == nil {
		conversation, err := h.getOne(r, existingID, currentUserID)

		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not load conversation")
			return
		}

		writeJSON(w, http.StatusOK, conversation)
		return
	}

	tx, err := h.DB.Begin(r.Context())

	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not start transaction")
		return
	}

	defer tx.Rollback(r.Context())

	var conversationID int64

	if err := tx.QueryRow(r.Context(), `
INSERT INTO conversations DEFAULT VALUES RETURNING id
`).Scan(&conversationID); err != nil {
		writeError(w, http.StatusInternalServerError, "could not create conversation")
		return
	}

	if _, err := tx.Exec(r.Context(), `
INSERT INTO conversation_participants (conversation_id, user_id)
VALUES ($1, $2), ($1, $3)
`, conversationID, currentUserID, input.UserID); err != nil {
		writeError(w, http.StatusBadRequest, "other user does not exist")
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "could not save conversation")
		return
	}

	conversation, err := h.getOne(r, conversationID, currentUserID)

	if err != nil {
		writeError(w, http.StatusInternalServerError, "conversation created but could not reload it")
		return
	}

	writeJSON(w, http.StatusCreated, conversation)
}

func (h ConversationHandler) List(w http.ResponseWriter, r *http.Request) {
	currentUserID := middleware.UserID(r)

	rows, err := h.DB.Query(r.Context(), `
SELECT c.id, c.created_at, u.id, u.name, u.email, u.created_at
FROM conversations c
JOIN conversation_participants mine
ON mine.conversation_id = c.id AND mine.user_id = $1
JOIN conversation_participants other
ON other.conversation_id = c.id AND other.user_id <> $1
JOIN users u ON u.id = other.user_id
ORDER BY c.id DESC
`, currentUserID)

	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load conversations")
		return
	}

	defer rows.Close()

	result := []models.Conversation{}

	for rows.Next() {
		var c models.Conversation

		if err := rows.Scan(
			&c.ID,
			&c.CreatedAt,
			&c.OtherUser.ID,
			&c.OtherUser.Name,
			&c.OtherUser.Email,
			&c.OtherUser.CreatedAt,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "could not read conversations")
			return
		}

		result = append(result, c)
	}

	writeJSON(w, http.StatusOK, result)
}

func (h ConversationHandler) getOne(
	r *http.Request,
	conversationID,
	currentUserID int64,
) (models.Conversation, error) {
	var c models.Conversation

	err := h.DB.QueryRow(r.Context(), `
SELECT c.id, c.created_at, u.id, u.name, u.email, u.created_at
FROM conversations c
JOIN conversation_participants mine
ON mine.conversation_id = c.id AND mine.user_id = $2
JOIN conversation_participants other
ON other.conversation_id = c.id AND other.user_id <> $2
JOIN users u ON u.id = other.user_id
WHERE c.id = $1
LIMIT 1
`, conversationID, currentUserID).Scan(
		&c.ID,
		&c.CreatedAt,
		&c.OtherUser.ID,
		&c.OtherUser.Name,
		&c.OtherUser.Email,
		&c.OtherUser.CreatedAt,
	)

	return c, err
}
