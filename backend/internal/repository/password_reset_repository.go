package repository

import (
	"context"
	"crypto/subtle"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrPasswordResetCooldown = errors.New("password reset code requested too recently")

type PasswordResetCode struct {
	ID        string
	UserID    string
	CodeHash  string
	ExpiresAt time.Time
	Attempts  int
	UsedAt    *time.Time
}

type PasswordResetRepository struct {
	db *pgxpool.Pool
}

func NewPasswordResetRepository(db *pgxpool.Pool) *PasswordResetRepository {
	return &PasswordResetRepository{db: db}
}

func (r *PasswordResetRepository) Create(
	ctx context.Context,
	userID string,
	codeHash string,
	expiresAt time.Time,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var lockedUserID string
	if err := tx.QueryRow(ctx, `
		SELECT id FROM users WHERE id = $1 FOR UPDATE
	`, userID).Scan(&lockedUserID); err != nil {
		return err
	}

	var recentlyRequested bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM password_reset_codes
			WHERE user_id = $1
			  AND created_at > NOW() - INTERVAL '1 minute'
		)
	`, userID).Scan(&recentlyRequested)
	if err != nil {
		return err
	}
	if recentlyRequested {
		return ErrPasswordResetCooldown
	}

	_, err = tx.Exec(ctx, `
		UPDATE password_reset_codes
		SET used_at = NOW()
		WHERE user_id = $1 AND used_at IS NULL
	`, userID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO password_reset_codes (user_id, code_hash, expires_at)
		VALUES ($1, $2, $3)
	`, userID, codeHash, expiresAt)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *PasswordResetRepository) GetActive(
	ctx context.Context,
	userID string,
) (*PasswordResetCode, error) {
	code := &PasswordResetCode{}
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, code_hash, expires_at, attempts, used_at
		FROM password_reset_codes
		WHERE user_id = $1 AND used_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`, userID).Scan(
		&code.ID,
		&code.UserID,
		&code.CodeHash,
		&code.ExpiresAt,
		&code.Attempts,
		&code.UsedAt,
	)
	if err != nil {
		return nil, err
	}
	return code, nil
}

func (r *PasswordResetRepository) IncrementFailedAttempt(
	ctx context.Context,
	codeID string,
) error {
	_, err := r.db.Exec(ctx, `
		UPDATE password_reset_codes
		SET attempts = attempts + 1,
		    used_at = CASE WHEN attempts + 1 >= 5 THEN NOW() ELSE used_at END
		WHERE id = $1 AND used_at IS NULL AND attempts < 5
	`, codeID)
	return err
}

func (r *PasswordResetRepository) Invalidate(
	ctx context.Context,
	codeID string,
) error {
	_, err := r.db.Exec(ctx, `
		UPDATE password_reset_codes
		SET used_at = NOW()
		WHERE id = $1 AND used_at IS NULL
	`, codeID)
	return err
}

func (r *PasswordResetRepository) Complete(
	ctx context.Context,
	userID string,
	codeID string,
	expectedCodeHash string,
	newPasswordHash string,
) (bool, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	var storedHash string
	var expiresAt time.Time
	var attempts int
	var usedAt *time.Time
	err = tx.QueryRow(ctx, `
		SELECT code_hash, expires_at, attempts, used_at
		FROM password_reset_codes
		WHERE id = $1 AND user_id = $2
		FOR UPDATE
	`, codeID, userID).Scan(&storedHash, &expiresAt, &attempts, &usedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if usedAt != nil || attempts >= 5 || !expiresAt.After(time.Now()) ||
		subtle.ConstantTimeCompare([]byte(storedHash), []byte(expectedCodeHash)) != 1 {
		return false, nil
	}

	commandTag, err := tx.Exec(ctx, `
		UPDATE users SET password_hash = $1 WHERE id = $2
	`, newPasswordHash, userID)
	if err != nil {
		return false, err
	}
	if commandTag.RowsAffected() != 1 {
		return false, nil
	}

	_, err = tx.Exec(ctx, `
		UPDATE password_reset_codes SET used_at = NOW() WHERE id = $1
	`, codeID)
	if err != nil {
		return false, err
	}

	_, err = tx.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE user_id = $1 AND revoked_at IS NULL
	`, userID)
	if err != nil {
		return false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return false, err
	}
	return true, nil
}
