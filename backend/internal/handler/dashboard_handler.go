package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"home-server-dashboard/backend/internal/middleware"
	"home-server-dashboard/backend/internal/service"

	"github.com/jackc/pgx/v5"
)

type DashboardHandler struct {
	service *service.DashboardService
}

func NewDashboardHandler(svc *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{
		service: svc,
	}
}

func (h *DashboardHandler) GetSummary(w http.ResponseWriter, r *http.Request) {

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	serverID := r.PathValue("server_id")

	summary, err := h.service.GetSummary(
		r.Context(),
		serverID,
		userID,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Server not found", http.StatusNotFound)
			return
		}

		http.Error(
			w,
			"Failed to retrieve dashboard summary",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(summary); err != nil {
		return
	}
}
