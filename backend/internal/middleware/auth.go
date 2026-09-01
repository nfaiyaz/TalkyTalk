package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userIDKey contextKey = "userID"

type Auth struct {
	Secret []byte
}

func (a Auth) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")

		if !strings.HasPrefix(header, "Bearer ") {
			writeUnauthorized(w)
			return
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))

		userID, err := a.ParseUserID(tokenString)
		if err != nil {
			writeUnauthorized(w)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a Auth) ParseUserID(tokenString string) (int64, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return a.Secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	if err != nil || !token.Valid {
		return 0, jwt.ErrTokenInvalidClaims
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, jwt.ErrTokenInvalidClaims
	}

	value, err := claims.GetSubject()
	if err != nil {
		return 0, err
	}

	var userID int64

	_, err = fmt.Sscan(value, &userID)

	return userID, err
}

func UserID(r *http.Request) int64 {
	value, _ := r.Context().Value(userIDKey).(int64)

	return value
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": "unauthorized",
	})
}


func fmtSscan(s string, value *int64) (int, error) {
	return fmt.Sscan(s, value)
}