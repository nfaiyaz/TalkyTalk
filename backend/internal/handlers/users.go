package handlers

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"go-chat-app/backend/internal/middleware"
	"go-chat-app/backend/internal/models"
)

type UserHandler struct {
	DB *pgxpool.Pool
}

func (h UserHandler) List(w http.ResponseWriter, r *http.Request) {
	currentUserID := middleware.UserID(r)

	rows, err := h.DB.Query(r.Context(), `
SELECT id, name, email, created_at
FROM users
WHERE id <> $1
ORDER BY name
`, currentUserID)

	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load users")
		return
	}

	defer rows.Close()

	users := []models.User{}

	for rows.Next() {
		var user models.User

		if err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.CreatedAt,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "could not read users")
			return
		}

		users = append(users, user)
	}

	writeJSON(w, http.StatusOK, users)
}