package service_test

import (
	"context"
	"errors"
	"testing"
	"weather-api/internal/dto"
	"weather-api/internal/model"
	"weather-api/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockUserCityRepository struct {
	mock.Mock
}

func (m *MockUserCityRepository) AddCity(ctx context.Context, city *model.UserCity) (model.UserCity, error) {
	args := m.Called(ctx, city)
	return args.Get(0).(model.UserCity), args.Error(1)
}

func (m *MockUserCityRepository) ListCities(ctx context.Context, userID int64) ([]model.UserCity, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.UserCity), args.Error(1)
}

func (m *MockUserCityRepository) DeleteCity(ctx context.Context, userID int64, cityID int64) error {
	args := m.Called(ctx, userID, cityID)
	return args.Error(0)
}

func TestUserCityService_AddCity_Success(t *testing.T) {
	userRepo := new(MockUserRepository)
	cityRepo := new(MockUserCityRepository)
	userService := service.NewUserService(userRepo)
	cityService := service.NewUserCityService(cityRepo, userService)

	userRepo.On("GetByID", mock.Anything, int64(1)).Return(model.User{ID: 1}, nil).Once()
	cityRepo.On("AddCity", mock.Anything, &model.UserCity{UserID: 1, City: "Almaty"}).Return(model.UserCity{ID: 10, UserID: 1, City: "Almaty"}, nil).Once()

	input := &dto.AddUserCityRequest{
		UserID: 1,
		City:   "Almaty",
	}
	res, err := cityService.AddCity(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, int64(10), res.ID)
	assert.Equal(t, "Almaty", res.City)
	userRepo.AssertExpectations(t)
	cityRepo.AssertExpectations(t)
}

func TestUserCityService_AddCity_EmptyCity(t *testing.T) {
	cityRepo := new(MockUserCityRepository)
	cityService := service.NewUserCityService(cityRepo, nil)

	input := &dto.AddUserCityRequest{
		UserID: 1,
		City:   "   ",
	}
	_, err := cityService.AddCity(context.Background(), input)

	require.ErrorIs(t, err, dto.ErrInvalidUserInput)
}

func TestUserCityService_AddCity_UserNotFound(t *testing.T) {
	userRepo := new(MockUserRepository)
	cityRepo := new(MockUserCityRepository)
	userService := service.NewUserService(userRepo)
	cityService := service.NewUserCityService(cityRepo, userService)

	userRepo.On("GetByID", mock.Anything, int64(1)).Return(model.User{}, model.ErrUserNotFound).Once()

	input := &dto.AddUserCityRequest{
		UserID: 1,
		City:   "Almaty",
	}
	_, err := cityService.AddCity(context.Background(), input)

	require.ErrorIs(t, err, model.ErrUserNotFound)
	userRepo.AssertExpectations(t)
}

func TestUserCityService_AddCity_AlreadyExists(t *testing.T) {
	userRepo := new(MockUserRepository)
	cityRepo := new(MockUserCityRepository)
	userService := service.NewUserService(userRepo)
	cityService := service.NewUserCityService(cityRepo, userService)

	userRepo.On("GetByID", mock.Anything, int64(1)).Return(model.User{ID: 1}, nil).Once()
	cityRepo.On("AddCity", mock.Anything, mock.Anything).Return(model.UserCity{}, errors.New("unique constraint violation: SQLSTATE 23505")).Once()

	input := &dto.AddUserCityRequest{
		UserID: 1,
		City:   "Almaty",
	}
	_, err := cityService.AddCity(context.Background(), input)

	require.ErrorIs(t, err, service.ErrCityAlreadyExists)
	userRepo.AssertExpectations(t)
	cityRepo.AssertExpectations(t)
}

func TestUserCityService_ListCities_Success(t *testing.T) {
	userRepo := new(MockUserRepository)
	cityRepo := new(MockUserCityRepository)
	userService := service.NewUserService(userRepo)
	cityService := service.NewUserCityService(cityRepo, userService)

	userRepo.On("GetByID", mock.Anything, int64(1)).Return(model.User{ID: 1}, nil).Once()
	cityRepo.On("ListCities", mock.Anything, int64(1)).Return([]model.UserCity{
		{ID: 10, City: "Almaty"},
	}, nil).Once()

	cities, err := cityService.ListCities(context.Background(), 1)

	require.NoError(t, err)
	assert.Len(t, cities, 1)
	assert.Equal(t, "Almaty", cities[0].City)
}

func TestUserCityService_DeleteCity_Success(t *testing.T) {
	userRepo := new(MockUserRepository)
	cityRepo := new(MockUserCityRepository)
	userService := service.NewUserService(userRepo)
	cityService := service.NewUserCityService(cityRepo, userService)

	userRepo.On("GetByID", mock.Anything, int64(1)).Return(model.User{ID: 1}, nil).Once()
	cityRepo.On("DeleteCity", mock.Anything, int64(1), int64(10)).Return(nil).Once()

	err := cityService.DeleteCity(context.Background(), 1, 10)

	require.NoError(t, err)
}
