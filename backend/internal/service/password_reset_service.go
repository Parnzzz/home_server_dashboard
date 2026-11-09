package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/mail"
	"strings"
	"time"

	"home-server-dashboard/backend/internal/repository"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	passwordResetCodeLength  = 8
	passwordResetMaxAttempts = 5
	passwordResetLifetime    = 15 * time.Minute
)

var (
	ErrPasswordResetUnavailable = errors.New("password reset email is not configured")
	ErrInvalidResetCode         = errors.New("invalid or expired reset code")
)

type PasswordResetUserRepository interface {
	GetUserByEmail(ctx context.Context, email string) (*repository.User, error)
}

type PasswordResetCodeRepository interface {
	Create(ctx context.Context, userID string, codeHash string, expiresAt time.Time) error
	GetActive(ctx context.Context, userID string) (*repository.PasswordResetCode, error)
	IncrementFailedAttempt(ctx context.Context, codeID string) error
	Invalidate(ctx context.Context, codeID string) error
	Complete(
		ctx context.Context,
		userID string,
		codeID string,
		expectedCodeHash string,
		newPasswordHash string,
	) (bool, error)
}

type PasswordResetService struct {
	users      PasswordResetUserRepository
	codes      PasswordResetCodeRepository
	mailer     PasswordResetMailer
	hashSecret []byte
}

func NewPasswordResetService(
	users PasswordResetUserRepository,
	codes PasswordResetCodeRepository,
	mailer PasswordResetMailer,
	secret string,
) *PasswordResetService {
	return &PasswordResetService{
		users:      users,
		codes:      codes,
		mailer:     mailer,
		hashSecret: []byte(secret),
	}
}

func (s *PasswordResetService) RequestCode(
	ctx context.Context,
	email string,
) error {
	if !s.mailer.Available() {
		return ErrPasswordResetUnavailable
	}

	email, err := normalizeResetEmail(email)
	if err != nil {
		return err
	}

	user, err := s.users.GetUserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}

	code, err := generatePasswordResetCode()
	if err != nil {
		return err
	}
	codeHash := s.hashCode(user.ID, code)
	if err := s.codes.Create(
		ctx,
		user.ID,
		codeHash,
		time.Now().Add(passwordResetLifetime),
	); err != nil {
		if errors.Is(err, repository.ErrPasswordResetCooldown) {
			return nil
		}
		return err
	}

	if err := s.mailer.SendCode(email, code); err != nil {
		if active, findErr := s.codes.GetActive(ctx, user.ID); findErr == nil {
			_ = s.codes.Invalidate(ctx, active.ID)
		}
		return err
	}
	return nil
}

func (s *PasswordResetService) ResetPassword(
	ctx context.Context,
	email string,
	code string,
	newPassword string,
) error {
	email, err := normalizeResetEmail(email)
	if err != nil || !isValidResetCode(code) {
		return ErrInvalidResetCode
	}
	if len(newPassword) < 8 || len([]byte(newPassword)) > 72 {
		return ErrInvalidInput
	}

	user, err := s.users.GetUserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrInvalidResetCode
	}
	if err != nil {
		return err
	}

	resetCode, err := s.codes.GetActive(ctx, user.ID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrInvalidResetCode
	}
	if err != nil {
		return err
	}
	if resetCode.UsedAt != nil ||
		resetCode.Attempts >= passwordResetMaxAttempts ||
		!resetCode.ExpiresAt.After(time.Now()) {
		return ErrInvalidResetCode
	}

	codeHash := s.hashCode(user.ID, code)
	if !hmac.Equal([]byte(codeHash), []byte(resetCode.CodeHash)) {
		if err := s.codes.IncrementFailedAttempt(ctx, resetCode.ID); err != nil {
			return err
		}
		return ErrInvalidResetCode
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(newPassword),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return err
	}

	completed, err := s.codes.Complete(
		ctx,
		user.ID,
		resetCode.ID,
		codeHash,
		string(passwordHash),
	)
	if err != nil {
		return err
	}
	if !completed {
		return ErrInvalidResetCode
	}
	return nil
}

func (s *PasswordResetService) hashCode(userID string, code string) string {
	hash := hmac.New(sha256.New, s.hashSecret)
	_, _ = fmt.Fprintf(hash, "%s:%s", userID, code)
	return hex.EncodeToString(hash.Sum(nil))
}

func normalizeResetEmail(email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return "", ErrInvalidInput
	}
	return email, nil
}

func isValidResetCode(code string) bool {
	if len(code) != passwordResetCodeLength {
		return false
	}
	for _, char := range code {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func generatePasswordResetCode() (string, error) {
	number, err := rand.Int(rand.Reader, big.NewInt(100_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%08d", number.Int64()), nil
}
