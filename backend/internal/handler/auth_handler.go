package handler

import (
	"encoding/json"
	"errors"
	"home-server-dashboard/backend/internal/middleware"
	"home-server-dashboard/backend/internal/service"
	"net/http"
	"strings"
	"time"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

type registerRequest struct {
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
}

type userResponse struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type loginResponse struct {
	User         userResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int64        `json:"expires_in"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		writeAuthJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid request body",
		})
		return
	}

	if strings.TrimSpace(req.Email) == "" ||
		strings.TrimSpace(req.DisplayName) == "" ||
		req.Password == "" {
		writeAuthJSON(w, http.StatusBadRequest, errorResponse{
			Error: "email, display_name and password are required",
		})
		return
	}

	user, err := h.authService.Register(
		r.Context(),
		req.Email,
		req.DisplayName,
		req.Password,
	)

	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput):
			writeAuthJSON(w, http.StatusBadRequest, errorResponse{
				Error: "invalid registration data",
			})

		case errors.Is(err, service.ErrEmailAlreadyExists):
			writeAuthJSON(w, http.StatusConflict, errorResponse{
				Error: "email already exists",
			})

		default:
			writeAuthJSON(w, http.StatusInternalServerError, errorResponse{
				Error: "internal server error",
			})
		}
		return
	}

	response := userResponse{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		Role:        user.Role,
		CreatedAt:   user.CreatedAt,
	}

	writeAuthJSON(w, http.StatusCreated, response)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		writeAuthJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid request body",
		})
		return
	}

	if strings.TrimSpace(req.Email) == "" || req.Password == "" {
		writeAuthJSON(w, http.StatusBadRequest, errorResponse{
			Error: "email and password are required",
		})
		return
	}

	auth, err := h.authService.Login(
		r.Context(),
		req.Email,
		req.Password,
	)

	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			writeAuthJSON(w, http.StatusUnauthorized, errorResponse{
				Error: "invalid email or password",
			})

		default:
			writeAuthJSON(w, http.StatusInternalServerError, errorResponse{
				Error: "internal server error",
			})
		}
		return
	}

	response := loginResponse{
		User: userResponse{
			ID:          auth.User.ID,
			Email:       auth.User.Email,
			DisplayName: auth.User.DisplayName,
			AvatarURL:   auth.User.AvatarURL,
			Role:        auth.User.Role,
			CreatedAt:   auth.User.CreatedAt,
		},
		AccessToken:  auth.AccessToken,
		RefreshToken: auth.RefreshToken,
		TokenType:    auth.TokenType,
		ExpiresIn:    auth.ExpiresIn,
	}

	writeAuthJSON(w, http.StatusOK, response)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil || strings.TrimSpace(req.RefreshToken) == "" {
		writeAuthJSON(w, http.StatusBadRequest, errorResponse{
			Error: "refresh_token is required",
		})
		return
	}

	auth, err := h.authService.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			writeAuthJSON(w, http.StatusUnauthorized, errorResponse{
				Error: "invalid or expired refresh token",
			})
			return
		}

		writeAuthJSON(w, http.StatusInternalServerError, errorResponse{
			Error: "internal server error",
		})
		return
	}

	writeAuthJSON(w, http.StatusOK, loginResponse{
		User: userResponse{
			ID:          auth.User.ID,
			Email:       auth.User.Email,
			DisplayName: auth.User.DisplayName,
			AvatarURL:   auth.User.AvatarURL,
			Role:        auth.User.Role,
			CreatedAt:   auth.User.CreatedAt,
		},
		AccessToken:  auth.AccessToken,
		RefreshToken: auth.RefreshToken,
		TokenType:    auth.TokenType,
		ExpiresIn:    auth.ExpiresIn,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req logoutRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil || strings.TrimSpace(req.RefreshToken) == "" {
		writeAuthJSON(w, http.StatusBadRequest, errorResponse{
			Error: "refresh_token is required",
		})
		return
	}

	if err := h.authService.Logout(r.Context(), req.RefreshToken); err != nil {
		writeAuthJSON(w, http.StatusInternalServerError, errorResponse{
			Error: "internal server error",
		})
		return
	}

	writeAuthJSON(w, http.StatusOK, map[string]string{
		"message": "logged out",
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeAuthJSON(w, http.StatusUnauthorized, errorResponse{
			Error: "unauthorized",
		})
		return
	}

	user, err := h.authService.GetCurrentUser(
		r.Context(),
		userID,
	)

	if err != nil {
		writeAuthJSON(w, http.StatusUnauthorized, errorResponse{
			Error: "unauthorized",
		})
		return
	}

	response := userResponse{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		Role:        user.Role,
		CreatedAt:   user.CreatedAt,
	}

	writeAuthJSON(w, http.StatusOK, response)
}

func writeAuthJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		return
	}
}
