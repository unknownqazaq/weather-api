package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"weather-api/internal/auth"
	"weather-api/internal/dto"
	"weather-api/internal/model"
	"weather-api/internal/service"
)

type UserCityService interface {
	AddCity(ctx context.Context, input *dto.AddUserCityRequest) (model.UserCity, error)
	ListCities(ctx context.Context, userID int64) ([]model.UserCity, error)
	DeleteCity(ctx context.Context, userID int64, cityID int64) error
}

type UserCityHandler struct {
	service *service.UserCityService
}

type citiesResponse struct {
	Data []dto.UserCityResponse `json:"data"`
}

func NewUserCityHandler(service *service.UserCityService) *UserCityHandler {
	return &UserCityHandler{service: service}
}

func (h *UserCityHandler) AddCity(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.UserClaimsFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}
	userID := claims.UserID

	var input dto.AddUserCityRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid json body"})
		return
	}
	input.UserID = userID

	city, err := h.service.AddCity(r.Context(), &input)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, dto.MapUserCity(city))
}

func (h *UserCityHandler) ListCities(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.UserClaimsFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}
	userID := claims.UserID

	cities, err := h.service.ListCities(r.Context(), userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	var mappedCities []dto.UserCityResponse
	for _, c := range cities {
		mappedCities = append(mappedCities, dto.MapUserCity(c))
	}

	writeJSON(w, http.StatusOK, citiesResponse{Data: mappedCities})
}

func (h *UserCityHandler) DeleteCity(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.UserClaimsFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}
	userID := claims.UserID

	cityID, err := parseIDParam(r, "city_id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid city id"})
		return
	}

	err = h.service.DeleteCity(r.Context(), userID, cityID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserCityHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, model.ErrInvalidUserID), errors.Is(err, dto.ErrInvalidUserInput):
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	case errors.Is(err, service.ErrCityAlreadyExists):
		writeJSON(w, http.StatusConflict, ErrorResponse{Error: err.Error()})
	case errors.Is(err, model.ErrUserNotFound):
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
}
