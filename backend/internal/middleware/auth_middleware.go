package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"home-server-dashboard/backend/internal/service"
)

type contextKey string

const userIDKey contextKey = "user_id"

type AuthMiddleware struct {
	tokenService *service.TokenService
}

func NewAuthMiddleware(
	tokenService *service.TokenService,
) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService: tokenService,
	}
}

// Authenticate validates the access token before calling the handler.
func (m *AuthMiddleware) Authenticate(
	next http.Handler,
) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		parts := strings.Fields(authHeader)

		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") {
			writeAuthError(w, http.StatusUnauthorized, "missing or invalid authorization header")
			return
		}

		tokenString := parts[1]

		claims, err := m.tokenService.ValidateAccessToken(tokenString)
		if err != nil {
			writeAuthError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		ctx := context.WithValue(
			r.Context(),
			userIDKey,
			claims.UserID,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// UserIDFromContext retrieves the authenticated user's ID.
func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}

func writeAuthError(
	w http.ResponseWriter,
	status int,
	message string,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
