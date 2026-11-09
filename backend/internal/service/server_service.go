package service

import (
	"context"
	"errors"
	"strings"

	"home-server-dashboard/backend/internal/repository"
)

type CreateServerInput struct {
	UserID   string
	Name     string
	Hostname string
	OS       string
}

type ServerService struct {
	repository *repository.ServerRepository
}

func NewServerService(repo *repository.ServerRepository) *ServerService {
	return &ServerService{
		repository: repo,
	}
}

func (s *ServerService) Create(
	ctx context.Context,
	input CreateServerInput,
) (*repository.Server, error) {

	input.Name = strings.TrimSpace(input.Name)
	input.Hostname = strings.TrimSpace(input.Hostname)
	input.OS = strings.TrimSpace(input.OS)

	if input.UserID == "" {
		return nil, errors.New("user_id is required")
	}

	if input.Name == "" {
		return nil, errors.New("server name is required")
	}

	if len(input.Name) > 100 {
		return nil, errors.New("server name must not exceed 100 characters")
	}

	return s.repository.Create(
		ctx,
		input.UserID,
		input.Name,
		input.Hostname,
		input.OS)
}
func (s *ServerService) GetAll(
	ctx context.Context,
	userID string,
) ([]repository.Server, error) {

	if userID == "" {
		return nil, errors.New("user_id is required")
	}

	return s.repository.FindAll(ctx, userID)
}

func (s *ServerService) GetByID(
	ctx context.Context,
	serverID string,
	userID string,
) (*repository.Server, error) {

	serverID = strings.TrimSpace(serverID)

	if serverID == "" {
		return nil, errors.New("server_id is required")
	}

	return s.repository.FindByID(ctx, serverID, userID)
}

func (s *ServerService) Delete(
	ctx context.Context,
	serverID string,
	userID string,
) (bool, error) {
	serverID = strings.TrimSpace(serverID)
	if serverID == "" {
		return false, errors.New("server_id is required")
	}
	if userID == "" {
		return false, errors.New("user_id is required")
	}

	return s.repository.Delete(ctx, serverID, userID)
}
