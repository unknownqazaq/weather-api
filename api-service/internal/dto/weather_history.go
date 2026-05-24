package dto

import (
	"strings"
	"time"
	"weather-api/internal/model"
)

type SaveWeatherHistoryRequest struct {
	UserID      int64   `json:"user_id"`
	City        string  `json:"city"`
	Temperature float64 `json:"temperature"`
	Description string  `json:"description"`
}

type WeatherHistoryFilter struct {
	City   string
	Limit  int
	Offset int
}

func (f *WeatherHistoryFilter) Normalize() {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	f.City = strings.TrimSpace(f.City)
}

type WeatherHistoryResponse struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	City        string    `json:"city"`
	Temperature float64   `json:"temperature"`
	Description string    `json:"description"`
	RequestedAt time.Time `json:"requested_at"`
}

func MapWeatherHistory(wh model.WeatherHistory) WeatherHistoryResponse {
	return WeatherHistoryResponse{
		ID:          wh.ID,
		UserID:      wh.UserID,
		City:        wh.City,
		Temperature: wh.Temperature,
		Description: wh.Description,
		RequestedAt: wh.RequestedAt,
	}
}
