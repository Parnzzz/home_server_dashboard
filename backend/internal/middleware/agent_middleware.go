package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"home-server-dashboard/backend/internal/service"
)

type agentContextKey string

const agentServerIDKey agentContextKey = "agent_server_id"

type AgentMiddleware struct {
	tokenService *service.AgentTokenService
}

func NewAgentMiddleware(
	tokenService *service.AgentTokenService,
) *AgentMiddleware {
	return &AgentMiddleware{
		tokenService: tokenService,
	}
}

func (m *AgentMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		parts := strings.Fields(authHeader)

		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			writeAgentAuthError(w)
			return
		}

		token := parts[1]

		agentToken, err := m.tokenService.Validate(
			r.Context(),
			token,
		)
		if err != nil {
			writeAgentAuthError(w)
			return
		}

		ctx := context.WithValue(
			r.Context(),
			agentServerIDKey,
			agentToken.ServerID,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AgentServerIDFromContext(ctx context.Context) (string, bool) {
	serverID, ok := ctx.Value(agentServerIDKey).(string)
	return serverID, ok
}

func writeAgentAuthError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": "Invalid or missing agent token",
	})
}
