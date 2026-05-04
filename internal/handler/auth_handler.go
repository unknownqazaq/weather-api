package handler

import (
	"encoding/json"
	"net/http"
	"weather-api/internal/domain"
	"weather-api/internal/service"
)

type AuthHandler struct {
	service *service.AuthService
}

func NewAuthHandler(service *service.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input domain.CreateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid json body"})
		return
	}

	user, err := h.service.Register(r.Context(), &input)
	if err != nil {
		switch err {
		case domain.ErrEmailAlreadyTaken:
			writeJSON(w, http.StatusConflict, ErrorResponse{Error: err.Error()})
		case domain.ErrInvalidUserInput:
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		}
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input service.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid json body"})
		return
	}

	authResp, err := h.service.Login(r.Context(), input)
	if err != nil {
		if err == domain.ErrUserNotFound {
			writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "invalid credentials"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, authResp)
}
