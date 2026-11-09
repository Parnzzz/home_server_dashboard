package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"home-server-dashboard/backend/internal/middleware"
	"home-server-dashboard/backend/internal/service"
)

type UserHandler struct {
	authService *service.AuthService
}

func NewUserHandler(authService *service.AuthService) *UserHandler {
	return &UserHandler{
		authService: authService,
	}
}

type updateProfileRequest struct {
	DisplayName string  `json:"display_name"`
	AvatarURL   *string `json:"avatar_url"`
}

func (h *UserHandler) UpdateProfile(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeAuthJSON(w, http.StatusUnauthorized, errorResponse{
			Error: "unauthorized",
		})
		return
	}

	var req updateProfileRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		writeAuthJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid request body",
		})
		return
	}

	if strings.TrimSpace(req.DisplayName) == "" {
		writeAuthJSON(w, http.StatusBadRequest, errorResponse{
			Error: "display_name is required",
		})
		return
	}

	user, err := h.authService.UpdateProfile(
		r.Context(),
		userID,
		req.DisplayName,
		req.AvatarURL,
	)

	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			writeAuthJSON(w, http.StatusBadRequest, errorResponse{
				Error: "invalid profile data",
			})
			return
		}

		writeAuthJSON(w, http.StatusInternalServerError, errorResponse{
			Error: "internal server error",
		})
		return
	}

	writeAuthJSON(w, http.StatusOK, userResponse{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		Role:        user.Role,
		CreatedAt:   user.CreatedAt,
	})
}
