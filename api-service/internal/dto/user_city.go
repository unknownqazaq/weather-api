package dto

import (
	"time"
	"weather-api/internal/model"
)

type AddUserCityRequest struct {
	UserID int64  `json:"-"`
	City   string `json:"city"`
}

type UserCityResponse struct {
	ID      int64     `json:"id"`
	UserID  int64     `json:"user_id"`
	City    string    `json:"city"`
	AddedAt time.Time `json:"added_at"`
}

func MapUserCity(uc model.UserCity) UserCityResponse {
	return UserCityResponse{
		ID:      uc.ID,
		UserID:  uc.UserID,
		City:    uc.City,
		AddedAt: uc.AddedAt,
	}
}
