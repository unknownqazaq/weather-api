package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"weather-api/internal/service"

	"github.com/go-chi/chi/v5"
)

type Service interface {
	GetWeather(ctx context.Context, lat, lon float64) (*service.WeatherResult, error)
	GetWeatherByCity(ctx context.Context, city string) (*service.CityWeather, error)
	GetWeatherByCountry(ctx context.Context, country string) (*service.CountryWeatherResult, error)
	GetTopWarmestCitiesByCountry(ctx context.Context, country string) (*service.TopWarmestResult, error)
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

func (h *WeatherHandler) GetWeatherByCity(w http.ResponseWriter, r *http.Request) {
	city := chi.URLParam(r, "city")

	result, err := h.service.GetWeatherByCity(r.Context(), city)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *WeatherHandler) GetWeatherByCountry(w http.ResponseWriter, r *http.Request) {
	country := chi.URLParam(r, "country")

	result, err := h.service.GetWeatherByCountry(r.Context(), country)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *WeatherHandler) GetTopWarmestCitiesByCountry(w http.ResponseWriter, r *http.Request) {
	country := chi.URLParam(r, "country")

	result, err := h.service.GetTopWarmestCitiesByCountry(r.Context(), country)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func writeServiceError(w http.ResponseWriter, err error) {
	if errors.Is(err, service.ErrCityNotFound) || errors.Is(err, service.ErrCountryNotFound) {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, `{"error":"failed to encode json"}`, http.StatusInternalServerError)
	}
}
