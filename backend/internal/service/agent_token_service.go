package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"

	"home-server-dashboard/backend/internal/repository"
)

var ErrInvalidAgentToken = errors.New("invalid agent token")

type AgentTokenService struct {
	repository *repository.AgentTokenRepository
}

func NewAgentTokenService(
	repo *repository.AgentTokenRepository,
) *AgentTokenService {
	return &AgentTokenService{
		repository: repo,
	}
}

// Generate creates a new random agent token and stores its hash.
func (s *AgentTokenService) Generate(
	ctx context.Context,
	serverID string,
) (string, error) {

	if serverID == "" {
		return "", errors.New("server_id is required")
	}

	// Generate 32 cryptographically secure random bytes.
	randomBytes := make([]byte, 32)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	// Encode the token so it can be sent in HTTP headers.
	token := base64.RawURLEncoding.EncodeToString(randomBytes)

	// Store only the hash in the database.
	tokenHash := hashAgentToken(token)

	if err := s.repository.Create(ctx, serverID, tokenHash); err != nil {
		return "", err
	}

	return token, nil
}

// Validate checks whether an agent token is valid.
func (s *AgentTokenService) Validate(
	ctx context.Context,
	token string,
) (*repository.AgentToken, error) {

	if token == "" {
		return nil, ErrInvalidAgentToken
	}

	tokenHash := hashAgentToken(token)

	agentToken, err := s.repository.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, ErrInvalidAgentToken
	}

	return agentToken, nil
}

// Revoke revokes the active agent token for a server.
func (s *AgentTokenService) Revoke(
	ctx context.Context,
	serverID string,
) error {

	if serverID == "" {
		return errors.New("server_id is required")
	}

	return s.repository.Revoke(ctx, serverID)
}

func hashAgentToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
