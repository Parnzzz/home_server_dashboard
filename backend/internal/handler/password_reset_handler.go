package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"home-server-dashboard/backend/internal/service"
)

type PasswordResetHandler struct {
	service *service.PasswordResetService
}

func NewPasswordResetHandler(svc *service.PasswordResetService) *PasswordResetHandler {
	return &PasswordResetHandler{service: svc}
}

type requestPasswordResetRequest struct {
	Email string `json:"email"`
}

type confirmPasswordResetRequest struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

const passwordResetRequestMessage = "If an account exists for this email, a password reset code has been sent."

func (h *PasswordResetHandler) RequestCode(w http.ResponseWriter, r *http.Request) {
	var req requestPasswordResetRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil || strings.TrimSpace(req.Email) == "" {
		writeAuthJSON(w, http.StatusBadRequest, errorResponse{
			Error: "email is required",
		})
		return
	}

	err := h.service.RequestCode(r.Context(), req.Email)
	if errors.Is(err, service.ErrInvalidInput) {
		writeAuthJSON(w, http.StatusBadRequest, errorResponse{
			Error: "enter a valid email address",
		})
		return
	}
	if errors.Is(err, service.ErrPasswordResetUnavailable) {
		writeAuthJSON(w, http.StatusServiceUnavailable, errorResponse{
			Error: "password reset email is not configured",
		})
		return
	}
	if err != nil {
		// Keep the response identical for known and unknown accounts.
		log.Printf("Password reset request could not be completed: %v", err)
	}

	writeAuthJSON(w, http.StatusAccepted, map[string]string{
		"message": passwordResetRequestMessage,
	})
}

func (h *PasswordResetHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	var req confirmPasswordResetRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeAuthJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid request body",
		})
		return
	}

	err := h.service.ResetPassword(
		r.Context(),
		req.Email,
		req.Code,
		req.NewPassword,
	)
	if errors.Is(err, service.ErrInvalidResetCode) {
		writeAuthJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid or expired reset code",
		})
		return
	}
	if errors.Is(err, service.ErrInvalidInput) {
		writeAuthJSON(w, http.StatusBadRequest, errorResponse{
			Error: "password must be 8 to 72 bytes long",
		})
		return
	}
	if err != nil {
		log.Printf("Password reset could not be completed: %v", err)
		writeAuthJSON(w, http.StatusInternalServerError, errorResponse{
			Error: "password reset could not be completed",
		})
		return
	}

	writeAuthJSON(w, http.StatusOK, map[string]string{
		"message": "password reset successfully",
	})
}
