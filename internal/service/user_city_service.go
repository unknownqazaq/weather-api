package service

import (
	"context"
	"errors"
	"strings"
	"weather-api/internal/dto"
	"weather-api/internal/model"
)

var (
	ErrCityAlreadyExists = errors.New("city is already tracked by the user")
)

type UserCityRepository interface {
	AddCity(ctx context.Context, city *model.UserCity) (model.UserCity, error)
	ListCities(ctx context.Context, userID int64) ([]model.UserCity, error)
	DeleteCity(ctx context.Context, userID int64, cityID int64) error
}

type UserCityService struct {
	repo        UserCityRepository
	userService *UserService
}

func NewUserCityService(repo UserCityRepository, userService *UserService) *UserCityService {
	return &UserCityService{
		repo:        repo,
		userService: userService,
	}
}

func (s *UserCityService) AddCity(ctx context.Context, input *dto.AddUserCityRequest) (model.UserCity, error) {

	input.City = strings.TrimSpace(input.City)
	if input.City == "" {
		return model.UserCity{}, dto.ErrInvalidUserInput
	}

	_, err := s.userService.GetByID(ctx, input.UserID)
	if err != nil {
		return model.UserCity{}, err
	}

	cityToCreate := &model.UserCity{
		UserID: input.UserID,
		City:   input.City,
	}

	city, err := s.repo.AddCity(ctx, cityToCreate)
	if err != nil {

		if strings.Contains(err.Error(), "unique constraint") || strings.Contains(err.Error(), "SQLSTATE 23505") {
			return model.UserCity{}, ErrCityAlreadyExists
		}
		return model.UserCity{}, err
	}

	return city, nil
}

func (s *UserCityService) ListCities(ctx context.Context, userID int64) ([]model.UserCity, error) {

	_, err := s.userService.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.repo.ListCities(ctx, userID)
}

func (s *UserCityService) DeleteCity(ctx context.Context, userID int64, cityID int64) error {

	_, err := s.userService.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	return s.repo.DeleteCity(ctx, userID, cityID)
}
