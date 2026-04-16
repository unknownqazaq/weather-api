package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"weather-api/internal/service"
)

type Service interface {
	GetWeather(ctx context.Context, lat, lon float64) (*service.WeatherResult, error)
}

type WeatherHandler struct {
	service Service
}

func NewWeatherHandler(service Service) *WeatherHandler {
	return &WeatherHandler{
		service: service,
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}


func (h *WeatherHandler) GetWeather(w http.ResponseWriter, r *http.Request) {
	latStr := r.URL.Query().Get("lat")
	lonStr := r.URL.Query().Get("lon")

	if latStr == "" || lonStr == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "query params lat and lon are required",
		})
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "invalid lat",
		})
		return
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "invalid lon",
		})
		return
	}

	result, err := h.service.GetWeather(r.Context(), lat, lon)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, `{"error":"failed to encode json"}`, http.StatusInternalServerError)
	}
}
