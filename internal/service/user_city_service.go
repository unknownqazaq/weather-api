package service

import (
	"context"
	"errors"
	"strings"
	"weather-api/internal/domain"
)

var (
	ErrCityAlreadyExists = errors.New("city is already tracked by the user")
)

type UserCityRepository interface {
	AddCity(ctx context.Context, input *domain.AddUserCityInput) (domain.UserCity, error)
	ListCities(ctx context.Context, userID int64) ([]domain.UserCity, error)
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

func (s *UserCityService) AddCity(ctx context.Context, input *domain.AddUserCityInput) (domain.UserCity, error) {

	input.City = strings.TrimSpace(input.City)
	if input.City == "" {
		return domain.UserCity{}, domain.ErrInvalidUserInput
	}

	_, err := s.userService.GetByID(ctx, input.UserID)
	if err != nil {
		return domain.UserCity{}, err
	}

	city, err := s.repo.AddCity(ctx, input)
	if err != nil {

		if strings.Contains(err.Error(), "unique constraint") || strings.Contains(err.Error(), "SQLSTATE 23505") {
			return domain.UserCity{}, ErrCityAlreadyExists
		}
		return domain.UserCity{}, err
	}

	return city, nil
}

func (s *UserCityService) ListCities(ctx context.Context, userID int64) ([]domain.UserCity, error) {

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
