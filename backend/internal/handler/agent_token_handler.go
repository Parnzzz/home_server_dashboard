package handler

import (
	"encoding/json"
	"net/http"

	"home-server-dashboard/backend/internal/middleware"
	"home-server-dashboard/backend/internal/service"
)

type AgentTokenHandler struct {
	serverService     *service.ServerService
	agentTokenService *service.AgentTokenService
}

type AgentTokenResponse struct {
	Token     string `json:"token"`
	TokenType string `json:"token_type"`
}

func NewAgentTokenHandler(
	serverService *service.ServerService,
	agentTokenService *service.AgentTokenService,
) *AgentTokenHandler {
	return &AgentTokenHandler{
		serverService:     serverService,
		agentTokenService: agentTokenService,
	}
}

// Generate issues or rotates an agent token for a server.
func (h *AgentTokenHandler) Generate(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	serverID := r.PathValue("server_id")

	// Verify that the server belongs to this user.
	_, err := h.serverService.GetByID(
		r.Context(),
		serverID,
		userID,
	)
	if err != nil {
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}

	token, err := h.agentTokenService.Generate(
		r.Context(),
		serverID,
	)
	if err != nil {
		http.Error(
			w,
			"Failed to generate agent token",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(AgentTokenResponse{
		Token:     token,
		TokenType: "Bearer",
	})
}

// Revoke revokes the agent token of a server.
func (h *AgentTokenHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	serverID := r.PathValue("server_id")

	// Verify server ownership before revoking.
	_, err := h.serverService.GetByID(
		r.Context(),
		serverID,
		userID,
	)
	if err != nil {
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}

	if err := h.agentTokenService.Revoke(
		r.Context(),
		serverID,
	); err != nil {
		http.Error(
			w,
			"Failed to revoke agent token",
			http.StatusInternalServerError,
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
