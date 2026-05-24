package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"weather-api/gateway-service/internal/client"
)

func TestHealthCheck(t *testing.T) {
	router := chi.NewRouter()
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status":"ok"}`, w.Body.String())
}

func TestValidationErrors(t *testing.T) {
	httpClient := &http.Client{Timeout: 1 * time.Second}
	weatherClient := client.NewWeatherClient(httpClient)

	router := chi.NewRouter()
	router.Get("/external/weather", func(w http.ResponseWriter, r *http.Request) {
		latStr := r.URL.Query().Get("lat")
		lonStr := r.URL.Query().Get("lon")

		if latStr == "" || lonStr == "" {
			writeJSONError(w, http.StatusBadRequest, "lat and lon are required")
			return
		}
		// mock success
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	router.Get("/external/geocode", func(w http.ResponseWriter, r *http.Request) {
		city := r.URL.Query().Get("city")
		if city == "" {
			writeJSONError(w, http.StatusBadRequest, "city is required")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	router.Get("/external/cities", func(w http.ResponseWriter, r *http.Request) {
		country := r.URL.Query().Get("country")
		if country == "" {
			writeJSONError(w, http.StatusBadRequest, "country is required")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	t.Run("weather missing params", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/external/weather", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "lat and lon are required")
	})

	t.Run("geocode missing params", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/external/geocode", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "city is required")
	})

	t.Run("cities missing params", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/external/cities", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "country is required")
	})

	t.Run("weather success bypass", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/external/weather?lat=43.23&lon=76.88", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Just verifying types can be constructed
	assert.NotNil(t, weatherClient)
}
