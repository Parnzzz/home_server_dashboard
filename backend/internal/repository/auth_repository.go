package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID           string
	Email        string
	DisplayName  string
	AvatarURL    *string
	Role         string
	PasswordHash string
	CreatedAt    time.Time
}

type RefreshToken struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

type AuthRepository struct {
	db *pgxpool.Pool
}

func NewAuthRepository(db *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{
		db: db,
	}
}

// CreateUser creates a new user account.
func (r *AuthRepository) CreateUser(
	ctx context.Context,
	email string,
	displayName string,
	passwordHash string,
) (*User, error) {

	query := `
		INSERT INTO users (
			email,
			display_name,
			password_hash
		)
		VALUES ($1, $2, $3)
		RETURNING
			id,
			email,
			display_name,
			avatar_url,
			role,
			password_hash,
			created_at
	`

	user := &User{}

	err := r.db.QueryRow(
		ctx,
		query,
		email,
		displayName,
		passwordHash,
	).Scan(
		&user.ID,
		&user.Email,
		&user.DisplayName,
		&user.AvatarURL,
		&user.Role,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByEmail finds a user by email.
func (r *AuthRepository) GetUserByEmail(
	ctx context.Context,
	email string,
) (*User, error) {

	query := `
		SELECT
			id,
			email,
			display_name,
			avatar_url,
            role,
			password_hash,
			created_at
		FROM users
		WHERE LOWER(email) = LOWER($1)
	`

	user := &User{}

	err := r.db.QueryRow(
		ctx,
		query,
		email,
	).Scan(
		&user.ID,
		&user.Email,
		&user.DisplayName,
		&user.AvatarURL,
		&user.Role,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}

		return nil, err
	}

	return user, nil
}

// GetUserByID finds a user by ID.
func (r *AuthRepository) GetUserByID(
	ctx context.Context,
	userID string,
) (*User, error) {

	query := `
		SELECT
			id,
			email,
			display_name,
			avatar_url,
            role,
			password_hash,
			created_at
		FROM users
		WHERE id = $1
	`

	user := &User{}

	err := r.db.QueryRow(
		ctx,
		query,
		userID,
	).Scan(
		&user.ID,
		&user.Email,
		&user.DisplayName,
		&user.AvatarURL,
		&user.Role,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}

		return nil, err
	}

	return user, nil
}
func (r *AuthRepository) CreateRefreshToken(
	ctx context.Context,
	userID string,
	tokenHash string,
	expiresAt time.Time,
) error {

	query := `
        INSERT INTO refresh_tokens (
            user_id,
            token_hash,
            expires_at
        )
        VALUES ($1, $2, $3)
    `

	_, err := r.db.Exec(
		ctx,
		query,
		userID,
		tokenHash,
		expiresAt,
	)

	return err
}

func (r *AuthRepository) GetRefreshToken(
	ctx context.Context,
	tokenHash string,
) (*RefreshToken, error) {

	query := `
        SELECT
            id,
            user_id,
            token_hash,
            expires_at,
            revoked_at,
            created_at
        FROM refresh_tokens
        WHERE token_hash = $1
    `

	token := &RefreshToken{}

	err := r.db.QueryRow(
		ctx,
		query,
		tokenHash,
	).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.RevokedAt,
		&token.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return token, nil
}

func (r *AuthRepository) RevokeRefreshToken(
	ctx context.Context,
	tokenHash string,
) error {

	query := `
        UPDATE refresh_tokens
        SET revoked_at = NOW()
        WHERE token_hash = $1
          AND revoked_at IS NULL
    `

	_, err := r.db.Exec(ctx, query, tokenHash)

	return err
}

// RotateRefreshToken revokes a valid refresh token and stores its replacement
// in a single transaction so the same token cannot be used more than once.
func (r *AuthRepository) RotateRefreshToken(
	ctx context.Context,
	oldTokenHash string,
	newTokenHash string,
	expiresAt time.Time,
) (string, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	var userID string
	err = tx.QueryRow(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE token_hash = $1
		  AND revoked_at IS NULL
		  AND expires_at > NOW()
		RETURNING user_id
	`, oldTokenHash).Scan(&userID)
	if err != nil {
		return "", err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO refresh_tokens (
			user_id,
			token_hash,
			expires_at
		)
		VALUES ($1, $2, $3)
	`, userID, newTokenHash, expiresAt)
	if err != nil {
		return "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}

	return userID, nil
}

func (r *AuthRepository) UpdateProfile(
	ctx context.Context,
	userID string,
	displayName string,
	avatarURL *string,
) (*User, error) {
	query := `
        UPDATE users
        SET
            display_name = $2,
            avatar_url = COALESCE($3, avatar_url)
        WHERE id = $1
        RETURNING
            id,
            email,
            display_name,
            avatar_url,
            role,
            password_hash,
            created_at
    `

	var user User

	err := r.db.QueryRow(
		ctx,
		query,
		userID,
		displayName,
		avatarURL,
	).Scan(
		&user.ID,
		&user.Email,
		&user.DisplayName,
		&user.AvatarURL,
		&user.Role,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}
