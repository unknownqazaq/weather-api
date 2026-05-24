package model

import "time"

type UserCity struct {
	ID      int64     `db:"id"`
	UserID  int64     `db:"user_id"`
	City    string    `db:"city"`
	AddedAt time.Time `db:"added_at"`
}
