package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MetricInput struct {
	ServerID         string
	CPUPercent       float64
	MemoryUsedBytes  int64
	MemoryTotalBytes int64
	DiskUsedBytes    int64
	DiskTotalBytes   int64
}

type Metric struct {
	ID               int64     `json:"id"`
	ServerID         string    `json:"server_id"`
	RecordedAt       time.Time `json:"recorded_at"`
	CPUPercent       float64   `json:"cpu_percent"`
	MemoryUsedBytes  int64     `json:"memory_used_bytes"`
	MemoryTotalBytes int64     `json:"memory_total_bytes"`
	DiskUsedBytes    int64     `json:"disk_used_bytes"`
	DiskTotalBytes   int64     `json:"disk_total_bytes"`
}

type MetricsRepository struct {
	db *pgxpool.Pool
}

func NewMetricsRepository(db *pgxpool.Pool) *MetricsRepository {
	return &MetricsRepository{
		db: db,
	}
}

func (r *MetricsRepository) Create(
	ctx context.Context,
	input MetricInput,
) error {
	query := `
		INSERT INTO metrics (
			server_id,
			cpu_percent,
			memory_used_bytes,
			memory_total_bytes,
			disk_used_bytes,
			disk_total_bytes
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		input.ServerID,
		input.CPUPercent,
		input.MemoryUsedBytes,
		input.MemoryTotalBytes,
		input.DiskUsedBytes,
		input.DiskTotalBytes,
	)

	return err
}

func (r *MetricsRepository) ListByServerID(
	ctx context.Context,
	serverID string,
	limit int,
) ([]Metric, error) {

	query := `
        SELECT
            id,
            server_id::text,
            recorded_at,
            cpu_percent,
            memory_used_bytes,
            memory_total_bytes,
            disk_used_bytes,
            disk_total_bytes
        FROM metrics
        WHERE server_id = $1
        ORDER BY recorded_at DESC
        LIMIT $2
    `

	rows, err := r.db.Query(ctx, query, serverID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	metrics := make([]Metric, 0)

	for rows.Next() {
		var metric Metric

		err := rows.Scan(
			&metric.ID,
			&metric.ServerID,
			&metric.RecordedAt,
			&metric.CPUPercent,
			&metric.MemoryUsedBytes,
			&metric.MemoryTotalBytes,
			&metric.DiskUsedBytes,
			&metric.DiskTotalBytes,
		)

		if err != nil {
			return nil, err
		}

		metrics = append(metrics, metric)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return metrics, nil
}

func (r *MetricsRepository) FindLatestByServerID(
	ctx context.Context,
	serverID string,
) (*Metric, error) {

	query := `
        SELECT
            id,
            server_id::text,
            recorded_at,
            cpu_percent,
            memory_used_bytes,
            memory_total_bytes,
            disk_used_bytes,
            disk_total_bytes
        FROM metrics
        WHERE server_id = $1
        ORDER BY recorded_at DESC
        LIMIT 1
    `

	var metric Metric

	err := r.db.QueryRow(
		ctx,
		query,
		serverID,
	).Scan(
		&metric.ID,
		&metric.ServerID,
		&metric.RecordedAt,
		&metric.CPUPercent,
		&metric.MemoryUsedBytes,
		&metric.MemoryTotalBytes,
		&metric.DiskUsedBytes,
		&metric.DiskTotalBytes,
	)

	if err != nil {
		return nil, err
	}

	return &metric, nil
}

func (r *MetricsRepository) ListByServerIDSince(
	ctx context.Context,
	serverID string,
	since time.Time,
	limit int,
) ([]Metric, error) {
	rows, err := r.db.Query(
		ctx,
		`
        SELECT
            id,
            server_id,
            recorded_at,
            cpu_percent,
            memory_used_bytes,
            memory_total_bytes,
            disk_used_bytes,
            disk_total_bytes
        FROM metrics
        WHERE server_id = $1
          AND recorded_at >= $2
        ORDER BY recorded_at DESC
        LIMIT $3
        `,
		serverID,
		since,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	metrics := make([]Metric, 0)

	for rows.Next() {
		var metric Metric

		err := rows.Scan(
			&metric.ID,
			&metric.ServerID,
			&metric.RecordedAt,
			&metric.CPUPercent,
			&metric.MemoryUsedBytes,
			&metric.MemoryTotalBytes,
			&metric.DiskUsedBytes,
			&metric.DiskTotalBytes,
		)
		if err != nil {
			return nil, err
		}

		metrics = append(metrics, metric)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return metrics, nil
}
