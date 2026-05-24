package handler

import (
	"encoding/json"
	"net/http"
	"weather-api/internal/dto"
	"weather-api/internal/model"
	"weather-api/internal/service"
)

type AuthHandler struct {
	service *service.AuthService
}

func NewAuthHandler(service *service.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input dto.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid json body"})
		return
	}

	userResponse, err := h.service.Register(r.Context(), &input)
	if err != nil {
		switch err {
		case model.ErrEmailAlreadyTaken:
			writeJSON(w, http.StatusConflict, ErrorResponse{Error: err.Error()})
		case dto.ErrInvalidUserInput:
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		}
		return
	}

	writeJSON(w, http.StatusCreated, userResponse)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid json body"})
		return
	}

	authResp, err := h.service.Login(r.Context(), input)
	if err != nil {
		if err == model.ErrUserNotFound {
			writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "invalid credentials"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, authResp)
}
