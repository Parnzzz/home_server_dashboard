package service

import (
	"context"
	"errors"
	"time"

	"home-server-dashboard/backend/internal/repository"

	"github.com/jackc/pgx/v5"
)

type DashboardSummary struct {
	Server           repository.Server `json:"server"`
	CPUPercent       *float64          `json:"cpu_percent,omitempty"`
	MemoryUsedBytes  *int64            `json:"memory_used_bytes,omitempty"`
	MemoryTotalBytes *int64            `json:"memory_total_bytes,omitempty"`
	DiskUsedBytes    *int64            `json:"disk_used_bytes,omitempty"`
	DiskTotalBytes   *int64            `json:"disk_total_bytes,omitempty"`
	LastSeenAt       *time.Time        `json:"last_seen_at,omitempty"`
}

type DashboardService struct {
	serverRepository  *repository.ServerRepository
	metricsRepository *repository.MetricsRepository
}

func NewDashboardService(
	serverRepo *repository.ServerRepository,
	metricsRepo *repository.MetricsRepository,
) *DashboardService {
	return &DashboardService{
		serverRepository:  serverRepo,
		metricsRepository: metricsRepo,
	}
}

func (s *DashboardService) GetSummary(
	ctx context.Context,
	serverID string,
	userID string,
) (*DashboardSummary, error) {

	if serverID == "" {
		return nil, errors.New("server_id is required")
	}

	server, err := s.serverRepository.FindByID(ctx, serverID, userID)
	if err != nil {
		return nil, err
	}

	summary := &DashboardSummary{
		Server: *server,
	}

	metric, err := s.metricsRepository.FindLatestByServerID(
		ctx,
		serverID,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		summary.Server.Status = "unknown"
		return summary, nil
	}

	if err != nil {
		return nil, err
	}

	summary.CPUPercent = &metric.CPUPercent
	summary.MemoryUsedBytes = &metric.MemoryUsedBytes
	summary.MemoryTotalBytes = &metric.MemoryTotalBytes
	summary.DiskUsedBytes = &metric.DiskUsedBytes
	summary.DiskTotalBytes = &metric.DiskTotalBytes
	summary.LastSeenAt = &metric.RecordedAt

	if time.Since(metric.RecordedAt) <= 30*time.Second {
		summary.Server.Status = "online"
	} else {
		summary.Server.Status = "offline"
	}

	return summary, nil
}
