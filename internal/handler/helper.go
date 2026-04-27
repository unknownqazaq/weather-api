package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"weather-api/internal/domain"

	"github.com/go-chi/chi/v5"
)

// ErrorResponse — стандартная структура для ошибок
type ErrorResponse struct {
	Error string `json:"error"`
}

// writeJSON отправляет JSON ответ с указанным статус-кодом
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// parseIDParam достает ID из параметров URL и проверяет его на валидность
func parseIDParam(r *http.Request, paramName string) (int64, error) {
	id, err := strconv.ParseInt(chi.URLParam(r, paramName), 10, 64)
	if err != nil || id <= 0 {
		return 0, domain.ErrInvalidUserID
	}
	return id, nil
}

// parseIntQuery читает int из query-параметров с fallback значением
func parseIntQuery(r *http.Request, key string, fallback int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return fallback
	}
	return parsed
}

// parseBoolQuery читает bool из query-параметров с fallback значением
func parseBoolQuery(r *http.Request, key string, fallback bool) bool {
	value := r.URL.Query().Get(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
