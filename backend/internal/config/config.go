package config

import (
	"fmt"
	"net/mail"
	"os"
	"strconv"
	"strings"
)

type Environment string

const (
	Development Environment = "development"
	Testing     Environment = "testing"
	Production  Environment = "production"
)

type Config struct {
	Environment        Environment
	Port               int
	DatabaseURL        string
	JWTSecret          string
	AccessTokenMinutes int
	RefreshTokenDays   int
	SMTPHost           string
	SMTPPort           int
	SMTPUsername       string
	SMTPPassword       string
	SMTPFrom           string
}

func Load() (*Config, error) {
	// Environment
	env := Environment(
		strings.ToLower(os.Getenv("APP_ENV")),
	)

	if env == "" {
		env = Development
	}

	switch env {
	case Development, Testing, Production:
	default:
		return nil, fmt.Errorf(
			"invalid APP_ENV: %s", env,
		)
	}

	// Port
	port := 8080

	if value := os.Getenv("PORT"); value != "" {
		parsedPort, err := strconv.Atoi(value)
		if err != nil || parsedPort < 1 || parsedPort > 65535 {
			return nil, fmt.Errorf("invalid PORT: %s", value)
		}

		port = parsedPort
	}

	// Database URL
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	//JWT

	jwtSecret := os.Getenv("JWT_SECRET")

	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	if len(jwtSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 bytes")
	}

	accessTokenMinutes := 15
	if value := os.Getenv("ACCESS_TOKEN_MINUTES"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 {
			return nil, fmt.Errorf("invalid ACCESS_TOKEN_MINUTES")
		}
		accessTokenMinutes = parsed
	}

	refreshTokenDays := 7
	if value := os.Getenv("REFRESH_TOKEN_DAYS"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 {
			return nil, fmt.Errorf("invalid REFRESH_TOKEN_DAYS")
		}
		refreshTokenDays = parsed
	}

	smtpHost := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	smtpPort := 587
	if value := os.Getenv("SMTP_PORT"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > 65535 {
			return nil, fmt.Errorf("invalid SMTP_PORT")
		}
		smtpPort = parsed
	}
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	smtpFrom := strings.TrimSpace(os.Getenv("SMTP_FROM"))
	if smtpHost != "" && smtpFrom == "" {
		return nil, fmt.Errorf("SMTP_FROM is required when SMTP_HOST is configured")
	}
	if smtpHost != "" {
		from, err := mail.ParseAddress(smtpFrom)
		if err != nil || from.Address != smtpFrom {
			return nil, fmt.Errorf("SMTP_FROM must be a valid email address")
		}
	}
	if (smtpUsername == "") != (smtpPassword == "") {
		return nil, fmt.Errorf("SMTP_USERNAME and SMTP_PASSWORD must be set together")
	}
	if smtpHost == "" && (smtpUsername != "" || smtpPassword != "" || smtpFrom != "") {
		return nil, fmt.Errorf("SMTP_HOST is required when SMTP settings are configured")
	}

	return &Config{
		Environment:        env,
		Port:               port,
		DatabaseURL:        databaseURL,
		JWTSecret:          jwtSecret,
		AccessTokenMinutes: accessTokenMinutes,
		RefreshTokenDays:   refreshTokenDays,
		SMTPHost:           smtpHost,
		SMTPPort:           smtpPort,
		SMTPUsername:       smtpUsername,
		SMTPPassword:       smtpPassword,
		SMTPFrom:           smtpFrom,
	}, nil
}
