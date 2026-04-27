package domain

import "time"

type UserCity struct {
	ID      int64     `db:"id" json:"id"`
	UserID  int64     `db:"user_id" json:"user_id"`
	City    string    `db:"city" json:"city"`
	AddedAt time.Time `db:"added_at" json:"added_at"`
}

type AddUserCityInput struct {
	UserID int64  `db:"user_id" json:"-"`
	City   string `db:"city" json:"city"`
}
