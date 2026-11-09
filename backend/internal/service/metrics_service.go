package service

import (
	"context"
	"errors"
	"home-server-dashboard/backend/internal/repository"
	"strings"
	"time"
)

type CreateMetricInput struct {
	ServerID         string
	CPUPercent       float64
	MemoryUsedBytes  int64
	MemoryTotalBytes int64
	DiskUsedBytes    int64
	DiskTotalBytes   int64
}

type MetricsService struct {
	serverRepository  *repository.ServerRepository
	metricsRepository *repository.MetricsRepository
}

func NewMetricsService(
	serverRepo *repository.ServerRepository,
	metricsRepo *repository.MetricsRepository) *MetricsService {
	return &MetricsService{
		serverRepository:  serverRepo,
		metricsRepository: metricsRepo,
	}
}

func (s *MetricsService) Create(
	ctx context.Context,
	input CreateMetricInput,
) error {

	input.ServerID = strings.TrimSpace(input.ServerID)

	if input.ServerID == "" {
		return errors.New("server_id is required")
	}

	if input.CPUPercent < 0 || input.CPUPercent > 100 {
		return errors.New("cpu_percent must be between 0 and 100")
	}

	if input.MemoryUsedBytes < 0 || input.MemoryTotalBytes <= 0 {
		return errors.New("invalid memory values")
	}

	if input.DiskUsedBytes < 0 || input.DiskTotalBytes <= 0 {
		return errors.New("invalid disk values")
	}

	if input.MemoryUsedBytes > input.MemoryTotalBytes {
		return errors.New("memory used cannot exceed memory total")
	}

	if input.DiskUsedBytes > input.DiskTotalBytes {
		return errors.New("disk used cannot exceed disk total")
	}

	return s.metricsRepository.Create(ctx, repository.MetricInput{
		ServerID:         input.ServerID,
		CPUPercent:       input.CPUPercent,
		MemoryUsedBytes:  input.MemoryUsedBytes,
		MemoryTotalBytes: input.MemoryTotalBytes,
		DiskUsedBytes:    input.DiskUsedBytes,
		DiskTotalBytes:   input.DiskTotalBytes,
	})
}

func (s *MetricsService) GetByServerID(
	ctx context.Context,
	serverID string,
	userID string,
	limit int,
) ([]repository.Metric, error) {
	serverID = strings.TrimSpace(serverID)
	userID = strings.TrimSpace(userID)

	if serverID == "" {
		return nil, errors.New("server_id is required")
	}
	if userID == "" {
		return nil, errors.New("user_id is required")
	}

	// Verify server ownership before retrieving metrics.
	_, err := s.serverRepository.FindByID(ctx, serverID, userID)

	if err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 20
	}

	if limit > 1000 {
		limit = 1000
	}

	return s.metricsRepository.ListByServerID(ctx, serverID, limit)
}

func (s *MetricsService) GetByServerIDSince(
	ctx context.Context,
	serverID string,
	userID string,
	duration time.Duration,
	limit int,
) ([]repository.Metric, error) {

	serverID = strings.TrimSpace(serverID)
	userID = strings.TrimSpace(userID)

	if serverID == "" {
		return nil, errors.New("server_id is required")
	}

	if userID == "" {
		return nil, errors.New("user_id is required")
	}

	if duration <= 0 {
		return nil, errors.New("duration must be greater than zero")
	}

	_, err := s.serverRepository.FindByID(ctx, serverID, userID)

	if err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 100
	}

	if limit > 1000 {
		limit = 1000
	}

	since := time.Now().Add(-duration)

	return s.metricsRepository.ListByServerIDSince(
		ctx,
		serverID,
		since,
		limit,
	)
}
