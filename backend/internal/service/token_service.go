package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("expired token")
)

type TokenService struct {
	secret             []byte
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
}

type AccessTokenClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func NewTokenService(
	secret string,
	accessTokenMinutes int,
	refreshTokenDays int,
) *TokenService {
	return &TokenService{
		secret:             []byte(secret),
		accessTokenExpiry:  time.Duration(accessTokenMinutes) * time.Minute,
		refreshTokenExpiry: time.Duration(refreshTokenDays) * 24 * time.Hour,
	}
}

// GenerateAccessToken creates a signed JWT.
func (s *TokenService) GenerateAccessToken(
	userID string,
) (string, error) {

	now := time.Now()

	claims := AccessTokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenExpiry)),
			Issuer:    "home-server-dashboard",
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString(s.secret)
}

// ValidateAccessToken validates JWT signature and expiration.
func (s *TokenService) ValidateAccessToken(
	tokenString string,
) (*AccessTokenClaims, error) {

	claims := &AccessTokenClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, ErrInvalidToken
			}

			return s.secret, nil
		},
		jwt.WithValidMethods([]string{
			jwt.SigningMethodHS256.Alg(),
		}),
		jwt.WithIssuer("home-server-dashboard"),
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}

		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// GenerateRefreshToken creates a cryptographically secure random token.
func (s *TokenService) GenerateRefreshToken() (string, error) {
	randomBytes := make([]byte, 32)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(randomBytes), nil
}

// HashRefreshToken hashes a refresh token using SHA-256.
func (s *TokenService) HashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// RefreshTokenExpiry returns the refresh token expiration time.
func (s *TokenService) RefreshTokenExpiry() time.Time {
	return time.Now().Add(s.refreshTokenExpiry)
}
