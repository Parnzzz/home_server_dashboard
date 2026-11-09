package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AgentToken struct {
	ID         string
	ServerID   string
	TokenHash  string
	CreatedAt  time.Time
	LastUsedAt *time.Time
	RevokedAt  *time.Time
}

type AgentTokenRepository struct {
	db *pgxpool.Pool
}

func NewAgentTokenRepository(db *pgxpool.Pool) *AgentTokenRepository {
	return &AgentTokenRepository{
		db: db,
	}
}

// Create creates an agent token for a server.
func (r *AgentTokenRepository) Create(
	ctx context.Context,
	serverID string,
	tokenHash string,
) error {
	query := `
		INSERT INTO agent_tokens (
			server_id,
			token_hash
		)
		VALUES ($1, $2)
		ON CONFLICT (server_id)
		DO UPDATE SET
			token_hash = EXCLUDED.token_hash,
			created_at = NOW(),
			last_used_at = NULL,
			revoked_at = NULL
	`

	_, err := r.db.Exec(
		ctx,
		query,
		serverID,
		tokenHash,
	)

	return err
}

// FindByTokenHash retrieves an active token.
func (r *AgentTokenRepository) FindByTokenHash(
	ctx context.Context,
	tokenHash string,
) (*AgentToken, error) {
	query := `
		SELECT
			id::text,
			server_id::text,
			token_hash,
			created_at,
			last_used_at,
			revoked_at
		FROM agent_tokens
		WHERE token_hash = $1
		  AND revoked_at IS NULL
	`

	var token AgentToken

	err := r.db.QueryRow(
		ctx,
		query,
		tokenHash,
	).Scan(
		&token.ID,
		&token.ServerID,
		&token.TokenHash,
		&token.CreatedAt,
		&token.LastUsedAt,
		&token.RevokedAt,
	)

	if err != nil {
		return nil, err
	}

	return &token, nil
}

// UpdateLastUsed updates the last usage timestamp.
func (r *AgentTokenRepository) UpdateLastUsed(
	ctx context.Context,
	tokenID string,
) error {
	query := `
		UPDATE agent_tokens
		SET last_used_at = NOW()
		WHERE id = $1
		  AND revoked_at IS NULL
	`

	_, err := r.db.Exec(ctx, query, tokenID)

	return err
}

// Revoke revokes an agent token.
func (r *AgentTokenRepository) Revoke(
	ctx context.Context,
	serverID string,
) error {
	query := `
		UPDATE agent_tokens
		SET revoked_at = NOW()
		WHERE server_id = $1
		  AND revoked_at IS NULL
	`

	_, err := r.db.Exec(ctx, query, serverID)

	return err
}
