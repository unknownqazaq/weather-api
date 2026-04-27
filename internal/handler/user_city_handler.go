package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"weather-api/internal/domain"
	"weather-api/internal/service"
)

type UserCityService interface {
	AddCity(ctx context.Context, input *domain.AddUserCityInput) (domain.UserCity, error)
	ListCities(ctx context.Context, userID int64) ([]domain.UserCity, error)
	DeleteCity(ctx context.Context, userID int64, cityID int64) error
}

type UserCityHandler struct {
	service *service.UserCityService
}

type citiesResponse struct {
	Data []domain.UserCity `json:"data"`
}

func NewUserCityHandler(service *service.UserCityService) *UserCityHandler {
	return &UserCityHandler{service: service}
}

func (h *UserCityHandler) AddCity(w http.ResponseWriter, r *http.Request) {
	userID, err := parseIDParam(r, "id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	var input domain.AddUserCityInput
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

	writeJSON(w, http.StatusCreated, city)
}

func (h *UserCityHandler) ListCities(w http.ResponseWriter, r *http.Request) {
	userID, err := parseIDParam(r, "id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	cities, err := h.service.ListCities(r.Context(), userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, citiesResponse{Data: cities})
}

func (h *UserCityHandler) DeleteCity(w http.ResponseWriter, r *http.Request) {
	userID, err := parseIDParam(r, "id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

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
	case errors.Is(err, domain.ErrInvalidUserID), errors.Is(err, domain.ErrInvalidUserInput):
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	case errors.Is(err, service.ErrCityAlreadyExists):
		writeJSON(w, http.StatusConflict, ErrorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrUserNotFound):
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
}
