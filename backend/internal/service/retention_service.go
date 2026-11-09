package service

import (
	"context"
	"errors"
	"log"
	"time"

	"home-server-dashboard/backend/internal/repository"
)

type RetentionCleaner interface {
	CleanupExpired(context.Context) (repository.RetentionCleanupCounts, error)
}

type RetentionService struct {
	cleaner RetentionCleaner
	logger  *log.Logger
}

func NewRetentionService(cleaner RetentionCleaner, logger *log.Logger) *RetentionService {
	return &RetentionService{
		cleaner: cleaner,
		logger:  logger,
	}
}

func (s *RetentionService) Run(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		s.clean(ctx)

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *RetentionService) clean(ctx context.Context) {
	cleanupContext, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	counts, err := s.cleaner.CleanupExpired(cleanupContext)
	if err != nil {
		if errors.Is(err, repository.ErrRetentionSchemaNotReady) {
			s.logger.Println("Database retention cleanup skipped; apply migrations 000005 and 000006 first")
			return
		}
		s.logger.Printf("Database retention cleanup failed after partial deletion: %v", err)
		return
	}

	s.logger.Printf(
		"Database retention cleanup completed: metrics=%d refresh_tokens=%d password_reset_codes=%d",
		counts.Metrics,
		counts.RefreshTokens,
		counts.ResetCodes,
	)
}
