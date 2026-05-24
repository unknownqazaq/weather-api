package service_test

import (
	"context"
	"testing"
	"weather-api/internal/dto"
	"weather-api/internal/model"
	"weather-api/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockWeatherHistoryRepository struct {
	mock.Mock
}

func (m *MockWeatherHistoryRepository) Save(ctx context.Context, history *model.WeatherHistory) (model.WeatherHistory, error) {
	args := m.Called(ctx, history)
	return args.Get(0).(model.WeatherHistory), args.Error(1)
}

func (m *MockWeatherHistoryRepository) GetHistory(ctx context.Context, userID int64, filter dto.WeatherHistoryFilter) ([]model.WeatherHistory, error) {
	args := m.Called(ctx, userID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.WeatherHistory), args.Error(1)
}

func TestUserWeatherService_GetUserWeather_Success(t *testing.T) {
	userRepo := new(MockUserRepository)
	cityRepo := new(MockUserCityRepository)
	provider := new(MockWeatherProvider)
	historyRepo := new(MockWeatherHistoryRepository)

	userService := service.NewUserService(userRepo)
	cityService := service.NewUserCityService(cityRepo, userService)
	weatherService := service.NewWeatherService(provider)
	userWeatherService := service.NewUserWeatherService(userService, cityService, weatherService, historyRepo)

	// Mock UserService.GetByID
	userRepo.On("GetByID", mock.Anything, int64(1)).Return(model.User{ID: 1}, nil).Twice() // called in userService and userCityService

	// Mock UserCityService.ListCities
	cityRepo.On("ListCities", mock.Anything, int64(1)).Return([]model.UserCity{
		{UserID: 1, City: "Almaty"},
	}, nil).Once()

	// Mock WeatherService.GetWeatherByCity
	provider.On("GetCityCoordinates", mock.Anything, "Almaty", "").Return(&service.ProviderCity{
		Name:      "Almaty",
		Country:   "Kazakhstan",
		Latitude:  43.2389,
		Longitude: 76.8897,
	}, nil).Once()
	provider.On("GetCurrentWeather", mock.Anything, 43.2389, 76.8897).Return(&service.ProviderWeatherResponse{
		Temperature: 15.0,
		WeatherCode: 1,
	}, nil).Once()

	// Mock HistoryRepo.Save
	historyRepo.On("Save", mock.Anything, mock.Anything).Return(model.WeatherHistory{}, nil).Once()

	res, err := userWeatherService.GetUserWeather(context.Background(), 1)

	require.NoError(t, err)
	assert.Equal(t, int64(1), res.UserID)
	require.Len(t, res.Cities, 1)
	assert.Equal(t, "Almaty", res.Cities[0].City)
	assert.Equal(t, 15.0, res.Cities[0].Temperature)

	userRepo.AssertExpectations(t)
	cityRepo.AssertExpectations(t)
	provider.AssertExpectations(t)
	historyRepo.AssertExpectations(t)
}

func TestUserWeatherService_GetUserWeather_EmptyCities(t *testing.T) {
	userRepo := new(MockUserRepository)
	cityRepo := new(MockUserCityRepository)
	historyRepo := new(MockWeatherHistoryRepository)

	userService := service.NewUserService(userRepo)
	cityService := service.NewUserCityService(cityRepo, userService)
	userWeatherService := service.NewUserWeatherService(userService, cityService, nil, historyRepo)

	userRepo.On("GetByID", mock.Anything, int64(1)).Return(model.User{ID: 1}, nil).Twice()
	cityRepo.On("ListCities", mock.Anything, int64(1)).Return([]model.UserCity{}, nil).Once()

	res, err := userWeatherService.GetUserWeather(context.Background(), 1)

	require.NoError(t, err)
	assert.Equal(t, int64(1), res.UserID)
	assert.Empty(t, res.Cities)
}

func TestUserWeatherService_GetHistory_Success(t *testing.T) {
	userRepo := new(MockUserRepository)
	historyRepo := new(MockWeatherHistoryRepository)

	userService := service.NewUserService(userRepo)
	userWeatherService := service.NewUserWeatherService(userService, nil, nil, historyRepo)

	userRepo.On("GetByID", mock.Anything, int64(1)).Return(model.User{ID: 1}, nil).Once()

	filter := dto.WeatherHistoryFilter{City: "Almaty"}
	historyRepo.On("GetHistory", mock.Anything, int64(1), mock.Anything).Return([]model.WeatherHistory{
		{ID: 100, UserID: 1, City: "Almaty", Temperature: 12.0},
	}, nil).Once()

	res, err := userWeatherService.GetHistory(context.Background(), 1, filter)

	require.NoError(t, err)
	assert.Equal(t, int64(1), res.UserID)
	assert.Equal(t, "Almaty", res.City)
	require.Len(t, res.History, 1)
	assert.Equal(t, 12.0, res.History[0].Temperature)
}
