package handler

import (
	"encoding/json"
	"home-server-dashboard/backend/internal/middleware"
	"home-server-dashboard/backend/internal/service"
	"net/http"
)

type ServerHandler struct {
	service *service.ServerService
}

type CreateServerRequest struct {
	Name     string `json:"name"`
	Hostname string `json:"hostname"`
	OS       string `json:"os"`
}

func NewServerHandler(svc *service.ServerService) *ServerHandler {
	return &ServerHandler{
		service: svc,
	}
}

func (h *ServerHandler) GetServers(w http.ResponseWriter, r *http.Request) {

	userID, ok := middleware.UserIDFromContext(r.Context())

	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	servers, err := h.service.GetAll(r.Context(), userID)
	if err != nil {
		http.Error(
			w,
			"Failed to retrieve servers",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(servers); err != nil {
		return
	}
}

func (h *ServerHandler) CreateServer(w http.ResponseWriter, r *http.Request) {

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var req CreateServerRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	server, err := h.service.Create(r.Context(), service.CreateServerInput{
		UserID:   userID,
		Name:     req.Name,
		Hostname: req.Hostname,
		OS:       req.OS,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(server); err != nil {
		return
	}
}

func (h *ServerHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	serverID := r.PathValue("server_id")

	server, err := h.service.GetByID(r.Context(), serverID, userID)
	if err != nil {
		http.Error(w, "Failed to retrieve server", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(server); err != nil {
		return
	}
}

func (h *ServerHandler) DeleteServer(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	deleted, err := h.service.Delete(
		r.Context(),
		r.PathValue("server_id"),
		userID,
	)
	if err != nil {
		http.Error(w, "Failed to delete server", http.StatusInternalServerError)
		return
	}
	if !deleted {
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
