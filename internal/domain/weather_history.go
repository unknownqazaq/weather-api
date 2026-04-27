package domain

import "time"

type WeatherHistory struct {
	ID          int64     `db:"id" json:"id"`
	UserID      int64     `db:"user_id" json:"user_id"`
	City        string    `db:"city" json:"city"`
	Temperature float64   `db:"temperature" json:"temperature"`
	Description string    `db:"description" json:"description"`
	RequestedAt time.Time `db:"requested_at" json:"requested_at"`
}

type SaveWeatherHistoryInput struct {
	UserID      int64   `db:"user_id"`
	City        string  `db:"city"`
	Temperature float64 `db:"temperature"`
	Description string  `db:"description"`
}

type WeatherHistoryFilter struct {
	City   string
	Limit  int
	Offset int
}
