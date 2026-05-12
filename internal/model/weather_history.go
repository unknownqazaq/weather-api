package model

import "time"

type WeatherHistory struct {
	ID          int64     `db:"id"`
	UserID      int64     `db:"user_id"`
	City        string    `db:"city"`
	Temperature float64   `db:"temperature"`
	Description string    `db:"description"`
	RequestedAt time.Time `db:"requested_at"`
}
