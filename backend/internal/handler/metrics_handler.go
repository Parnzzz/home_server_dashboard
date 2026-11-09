package handler

import (
	"encoding/json"
	"errors"
	"home-server-dashboard/backend/internal/middleware"
	"home-server-dashboard/backend/internal/repository"
	"home-server-dashboard/backend/internal/service"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

type MetricsHandler struct {
	service *service.MetricsService
}

func NewMetricsHandler(svc *service.MetricsService) *MetricsHandler {
	return &MetricsHandler{
		service: svc,
	}
}

type CreateMetricRequest struct {
	CPUPercent       float64 `json:"cpu_percent"`
	MemoryUsedBytes  int64   `json:"memory_used_bytes"`
	MemoryTotalBytes int64   `json:"memory_total_bytes"`
	DiskUsedBytes    int64   `json:"disk_used_bytes"`
	DiskTotalBytes   int64   `json:"disk_total_bytes"`
}

func (h *MetricsHandler) Create(w http.ResponseWriter, r *http.Request) {

	serverID, ok := middleware.AgentServerIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateMetricRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.service.Create(r.Context(), service.CreateMetricInput{
		ServerID:         serverID,
		CPUPercent:       req.CPUPercent,
		MemoryUsedBytes:  req.MemoryUsedBytes,
		MemoryTotalBytes: req.MemoryTotalBytes,
		DiskUsedBytes:    req.DiskUsedBytes,
		DiskTotalBytes:   req.DiskTotalBytes,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *MetricsHandler) GetByServerID(w http.ResponseWriter, r *http.Request) {

	userID, ok := middleware.UserIDFromContext(r.Context())

	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	serverID := r.PathValue("server_id")

	limit := 100
	if value := r.URL.Query().Get("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			http.Error(w, "Invalid limit", http.StatusBadRequest)
			return
		}
		limit = parsed
	}

	from := r.URL.Query().Get("from")

	var metrics []repository.Metric
	var err error

	if from != "" {
		duration, parseErr := time.ParseDuration(from)
		if parseErr != nil || duration <= 0 {
			http.Error(w, "Invalid from parameter", http.StatusBadRequest)
			return
		}

		metrics, err = h.service.GetByServerIDSince(
			r.Context(),
			serverID,
			userID,
			duration,
			limit,
		)
	} else {
		metrics, err = h.service.GetByServerID(
			r.Context(),
			serverID,
			userID,
			limit,
		)
	}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Server not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Failed to retrieve metrics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		return
	}
}
