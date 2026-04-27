package handler

import (
	"context"
	"errors"
	"net/http"
	"weather-api/internal/domain"
	"weather-api/internal/service"
)

type UserWeatherService interface {
	GetUserWeather(ctx context.Context, userID int64) (*service.UserWeatherResult, error)
	GetHistory(ctx context.Context, userID int64, filter domain.WeatherHistoryFilter) (*service.HistoryResponse, error)
}

type UserWeatherHandler struct {
	service *service.UserWeatherService
}

func NewUserWeatherHandler(service *service.UserWeatherService) *UserWeatherHandler {
	return &UserWeatherHandler{service: service}
}

func (h *UserWeatherHandler) GetWeather(w http.ResponseWriter, r *http.Request) {
	userID, err := parseIDParam(r, "id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	result, err := h.service.GetUserWeather(r.Context(), userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *UserWeatherHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	userID, err := parseIDParam(r, "id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	filter := domain.WeatherHistoryFilter{
		City:   r.URL.Query().Get("city"),
		Limit:  parseIntQuery(r, "limit", 0),
		Offset: parseIntQuery(r, "offset", 0),
	}

	result, err := h.service.GetHistory(r.Context(), userID, filter)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *UserWeatherHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidUserID):
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrUserNotFound):
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error: " + err.Error()})
	}
}
