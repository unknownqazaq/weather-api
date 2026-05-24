package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"weather-api/gateway-service/internal/client"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	port := getEnv("APP_PORT", "8081")
	apiServiceURL := getEnv("API_SERVICE_URL", "http://localhost:8080")

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	weatherClient := client.NewWeatherClient(httpClient)

	targetURL, err := url.Parse(apiServiceURL)
	if err != nil {
		logger.Fatal("invalid api service url", zap.Error(err), zap.String("url", apiServiceURL))
	}
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	router := chi.NewRouter()

	// Simple structured logger middleware
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := &statusResponseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(ww, r)
			logger.Info("request processed",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", ww.status),
				zap.Duration("duration", time.Since(start)),
			)
		})
	})

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	router.Get("/external/weather", func(w http.ResponseWriter, r *http.Request) {
		latStr := r.URL.Query().Get("lat")
		lonStr := r.URL.Query().Get("lon")

		if latStr == "" || lonStr == "" {
			writeJSONError(w, http.StatusBadRequest, "lat and lon are required")
			return
		}

		lat, err := strconv.ParseFloat(latStr, 64)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid lat")
			return
		}

		lon, err := strconv.ParseFloat(lonStr, 64)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid lon")
			return
		}

		resp, err := weatherClient.GetCurrentWeather(r.Context(), lat, lon)
		if err != nil {
			logger.Error("failed to get current weather", zap.Error(err))
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeJSON(w, http.StatusOK, resp)
	})

	router.Get("/external/geocode", func(w http.ResponseWriter, r *http.Request) {
		city := r.URL.Query().Get("city")
		country := r.URL.Query().Get("country")

		if city == "" {
			writeJSONError(w, http.StatusBadRequest, "city is required")
			return
		}

		resp, err := weatherClient.GetCityCoordinates(r.Context(), city, country)
		if err != nil {
			logger.Error("failed to get city coordinates", zap.Error(err))
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if resp == nil {
			writeJSONError(w, http.StatusNotFound, "city not found")
			return
		}

		writeJSON(w, http.StatusOK, resp)
	})

	router.Get("/external/cities", func(w http.ResponseWriter, r *http.Request) {
		country := r.URL.Query().Get("country")
		limitStr := r.URL.Query().Get("limit")

		if country == "" {
			writeJSONError(w, http.StatusBadRequest, "country is required")
			return
		}

		limit := 10
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil {
				limit = l
			}
		}

		resp, err := weatherClient.GetCitiesByCountry(r.Context(), country, limit)
		if err != nil {
			logger.Error("failed to get cities by country", zap.Error(err))
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if resp == nil {
			writeJSONError(w, http.StatusNotFound, "country or cities not found")
			return
		}

		writeJSON(w, http.StatusOK, resp)
	})

	// Fallback catch-all: reverse proxy to API Service
	router.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})

	addr := ":" + port
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("gateway server started", zap.String("addr", addr), zap.String("proxy_target", apiServiceURL))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("gateway listen server", zap.Error(err))
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("gateway shutdown server error", zap.Error(err))
	}

	logger.Info("gateway server stopped gracefully")
}

type statusResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
