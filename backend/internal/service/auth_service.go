package service

import (
	"context"
	"errors"
	"home-server-dashboard/backend/internal/repository"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidInput       = errors.New("invalid input")
)

type AuthRepository interface {
	CreateUser(
		ctx context.Context,
		email string,
		displayName string,
		passwordHash string,
	) (*repository.User, error)

	GetUserByEmail(
		ctx context.Context,
		email string,
	) (*repository.User, error)

	GetUserByID(
		ctx context.Context,
		userID string,
	) (*repository.User, error)

	CreateRefreshToken(
		ctx context.Context,
		userID string,
		tokenHash string,
		expiresAt time.Time,
	) error

	GetRefreshToken(
		ctx context.Context,
		tokenHash string,
	) (*repository.RefreshToken, error)

	RevokeRefreshToken(
		ctx context.Context,
		tokenHash string,
	) error

	RotateRefreshToken(
		ctx context.Context,
		oldTokenHash string,
		newTokenHash string,
		expiresAt time.Time,
	) (string, error)

	UpdateProfile(
		ctx context.Context,
		userID string,
		displayName string,
		avatarURL *string,
	) (*repository.User, error)
}

type AuthService struct {
	repo         AuthRepository
	tokenService *TokenService
}

type AuthResponse struct {
	User         *repository.User
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int64
}

func NewAuthService(repo AuthRepository, tokenService *TokenService) *AuthService {
	return &AuthService{
		repo:         repo,
		tokenService: tokenService,
	}
}

// Register creates a new user account.
func (s *AuthService) Register(
	ctx context.Context,
	email string,
	displayName string,
	password string,
) (*repository.User, error) {

	email = strings.ToLower(strings.TrimSpace(email))
	displayName = strings.TrimSpace(displayName)

	if email == "" || displayName == "" || password == "" {
		return nil, ErrInvalidInput
	}

	if len(password) < 8 {
		return nil, ErrInvalidInput
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.CreateUser(
		ctx,
		email,
		displayName,
		string(passwordHash),
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrEmailAlreadyExists
		}

		return nil, err
	}

	return user, nil
}

// Login authenticates a user using email and password.
func (s *AuthService) Login(
	ctx context.Context,
	email string,
	password string,
) (*AuthResponse, error) {

	email = strings.ToLower(strings.TrimSpace(email))

	if email == "" || password == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}

		return nil, err
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	)

	if err != nil {
		return nil, ErrInvalidCredentials
	}
	// 3. Generate Access Token
	accessToken, err := s.tokenService.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, err
	}

	// 4. Generate Refresh Token
	refreshToken, err := s.tokenService.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	// 5. Hash Refresh Token
	tokenHash := s.tokenService.HashRefreshToken(refreshToken)

	// 6. Calculate expiration
	expiresAt := s.tokenService.RefreshTokenExpiry()

	// 7. Save Refresh Token Hash
	err = s.repo.CreateRefreshToken(
		ctx,
		user.ID,
		tokenHash,
		expiresAt,
	)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.tokenService.accessTokenExpiry.Seconds()),
	}, nil
}

// Refresh validates and rotates a refresh token, returning a new token pair.
func (s *AuthService) Refresh(
	ctx context.Context,
	refreshToken string,
) (*AuthResponse, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, ErrInvalidToken
	}

	oldTokenHash := s.tokenService.HashRefreshToken(refreshToken)
	storedToken, err := s.repo.GetRefreshToken(ctx, oldTokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}
	if storedToken.RevokedAt != nil || !storedToken.ExpiresAt.After(time.Now()) {
		return nil, ErrInvalidToken
	}

	user, err := s.repo.GetUserByID(ctx, storedToken.UserID)
	if err != nil {
		return nil, err
	}

	accessToken, err := s.tokenService.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.tokenService.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	newRefreshTokenHash := s.tokenService.HashRefreshToken(newRefreshToken)
	userID, err := s.repo.RotateRefreshToken(
		ctx,
		oldTokenHash,
		newRefreshTokenHash,
		s.tokenService.RefreshTokenExpiry(),
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}
	if userID != user.ID {
		return nil, ErrInvalidToken
	}

	return &AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.tokenService.accessTokenExpiry.Seconds()),
	}, nil
}

// Logout revokes the current refresh token. Repeating the request is safe.
func (s *AuthService) Logout(
	ctx context.Context,
	refreshToken string,
) error {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return ErrInvalidToken
	}

	tokenHash := s.tokenService.HashRefreshToken(refreshToken)
	return s.repo.RevokeRefreshToken(ctx, tokenHash)
}

func (s *AuthService) GetCurrentUser(
	ctx context.Context,
	userID string,
) (*repository.User, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrInvalidInput
	}

	return s.repo.GetUserByID(ctx, userID)
}

func (s *AuthService) UpdateProfile(
	ctx context.Context,
	userID string,
	displayName string,
	avatarURL *string,
) (*repository.User, error) {
	displayName = strings.TrimSpace(displayName)

	if userID == "" || displayName == "" {
		return nil, ErrInvalidInput
	}

	if len(displayName) > 100 {
		return nil, ErrInvalidInput
	}

	if avatarURL != nil {
		value := strings.TrimSpace(*avatarURL)
		avatarURL = &value

		if value == "" {
			return nil, ErrInvalidInput
		}
	}

	return s.repo.UpdateProfile(
		ctx,
		userID,
		displayName,
		avatarURL,
	)
}

//GetUserByEmail

//ค้นหาบัญชีใน PostgreSQL

//CompareHashAndPassword

//ตรวจสอบ Password ด้วย bcrypt

//GenerateAccessToken

//สร้าง JWT อายุ 15 นาที

//GenerateRefreshToken

//สุ่ม Token จำนวน 32 bytes

//HashRefreshToken

//Hash ด้วย SHA-256

//CreateRefreshToken

//บันทึก Hash ลง PostgreSQL โดยมีอายุ 7 วัน

//Return AuthResponse

//ส่ง Token กลับให้ Handler
