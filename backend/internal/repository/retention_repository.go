package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const retentionBatchSize = 2000

var ErrRetentionSchemaNotReady = errors.New("retention migrations are not applied")

type RetentionCleanupCounts struct {
	Metrics       int64
	RefreshTokens int64
	ResetCodes    int64
}

type RetentionRepository struct {
	db *pgxpool.Pool
}

func NewRetentionRepository(db *pgxpool.Pool) *RetentionRepository {
	return &RetentionRepository{db: db}
}

func (r *RetentionRepository) CleanupExpired(ctx context.Context) (RetentionCleanupCounts, error) {
	var counts RetentionCleanupCounts
	var err error
	var schemaReady bool
	err = r.db.QueryRow(ctx, `
		SELECT
			to_regclass('metrics') IS NOT NULL
			AND to_regclass('refresh_tokens') IS NOT NULL
			AND to_regclass('password_reset_codes') IS NOT NULL
			AND to_regclass('idx_metrics_recorded_at') IS NOT NULL
			AND to_regclass('idx_refresh_tokens_expires_at') IS NOT NULL
			AND to_regclass('idx_refresh_tokens_revoked_at') IS NOT NULL
			AND to_regclass('idx_password_reset_codes_created_at') IS NOT NULL
	`).Scan(&schemaReady)
	if err != nil {
		return counts, err
	}
	if !schemaReady {
		return counts, ErrRetentionSchemaNotReady
	}

	counts.Metrics, err = deleteInBatches(ctx, r.db, `
		WITH expired AS (
			SELECT id
			FROM metrics
			WHERE recorded_at < NOW() - INTERVAL '7 days'
			ORDER BY recorded_at
			LIMIT $1
		)
		DELETE FROM metrics AS metric
		USING expired
		WHERE metric.id = expired.id
	`)
	if err != nil {
		return counts, err
	}

	counts.RefreshTokens, err = deleteInBatches(ctx, r.db, `
		WITH expired AS (
			SELECT id
			FROM refresh_tokens
			WHERE (revoked_at IS NOT NULL AND revoked_at < NOW() - INTERVAL '30 days')
			   OR expires_at < NOW() - INTERVAL '30 days'
			ORDER BY created_at
			LIMIT $1
		)
		DELETE FROM refresh_tokens AS token
		USING expired
		WHERE token.id = expired.id
	`)
	if err != nil {
		return counts, err
	}

	counts.ResetCodes, err = deleteInBatches(ctx, r.db, `
		WITH expired AS (
			SELECT id
			FROM password_reset_codes
			WHERE created_at < NOW() - INTERVAL '1 day'
			ORDER BY created_at
			LIMIT $1
		)
		DELETE FROM password_reset_codes AS code
		USING expired
		WHERE code.id = expired.id
	`)
	if err != nil {
		return counts, err
	}

	return counts, nil
}

func deleteInBatches(
	ctx context.Context,
	db *pgxpool.Pool,
	query string,
) (int64, error) {
	var total int64

	for {
		result, err := db.Exec(ctx, query, retentionBatchSize)
		if err != nil {
			return total, err
		}

		deleted := result.RowsAffected()
		total += deleted
		if deleted < retentionBatchSize {
			return total, nil
		}

		select {
		case <-ctx.Done():
			return total, ctx.Err()
		case <-time.After(25 * time.Millisecond):
		}
	}
}
