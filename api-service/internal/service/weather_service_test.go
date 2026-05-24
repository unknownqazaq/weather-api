package service_test

import (
	"context"
	"testing"
	"weather-api/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockWeatherProvider struct {
	mock.Mock
}

func (m *MockWeatherProvider) GetCurrentWeather(ctx context.Context, lat, lon float64) (*service.ProviderWeatherResponse, error) {
	args := m.Called(ctx, lat, lon)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.ProviderWeatherResponse), args.Error(1)
}

func (m *MockWeatherProvider) GetCityCoordinates(ctx context.Context, city, country string) (*service.ProviderCity, error) {
	args := m.Called(ctx, city, country)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.ProviderCity), args.Error(1)
}

func (m *MockWeatherProvider) GetCitiesByCountry(ctx context.Context, country string, limit int) ([]service.ProviderCity, error) {
	args := m.Called(ctx, country, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]service.ProviderCity), args.Error(1)
}

func TestWeatherService_GetWeather_Success(t *testing.T) {
	provider := new(MockWeatherProvider)
	weatherService := service.NewWeatherService(provider)

	provider.On("GetCurrentWeather", mock.Anything, 43.2389, 76.8897).Return(&service.ProviderWeatherResponse{
		Temperature: 15.5,
		WindSpeed:   3.2,
		WeatherCode: 1,
		Time:        "2026-05-20T12:00:00Z",
	}, nil).Once()

	res, err := weatherService.GetWeather(context.Background(), 43.2389, 76.8897)

	require.NoError(t, err)
	assert.Equal(t, 15.5, res.Temperature)
	assert.Equal(t, "Переменная облачность", res.Description)
	assert.Equal(t, "Куртка", res.Outfit)
	provider.AssertExpectations(t)
}

func TestWeatherService_GetWeather_Error(t *testing.T) {
	provider := new(MockWeatherProvider)
	weatherService := service.NewWeatherService(provider)

	provider.On("GetCurrentWeather", mock.Anything, 43.2389, 76.8897).Return(nil, assert.AnError).Once()

	res, err := weatherService.GetWeather(context.Background(), 43.2389, 76.8897)

	require.Error(t, err)
	assert.Nil(t, res)
	provider.AssertExpectations(t)
}

func TestWeatherService_GetWeatherByCity_Success(t *testing.T) {
	provider := new(MockWeatherProvider)
	weatherService := service.NewWeatherService(provider)

	provider.On("GetCityCoordinates", mock.Anything, "Almaty", "").Return(&service.ProviderCity{
		Name:      "Almaty",
		Country:   "Kazakhstan",
		Latitude:  43.2389,
		Longitude: 76.8897,
	}, nil).Once()

	provider.On("GetCurrentWeather", mock.Anything, 43.2389, 76.8897).Return(&service.ProviderWeatherResponse{
		Temperature: 25.0,
		WindSpeed:   1.5,
		WeatherCode: 0,
		Time:        "2026-05-20T12:00:00Z",
	}, nil).Once()

	res, err := weatherService.GetWeatherByCity(context.Background(), "Almaty")

	require.NoError(t, err)
	assert.Equal(t, "Almaty", res.City)
	assert.Equal(t, "Kazakhstan", res.Country)
	assert.Equal(t, 25.0, res.Temperature)
	assert.Equal(t, "Ясно", res.Description)
	assert.Equal(t, "Лёгкая одежда", res.Outfit)
	provider.AssertExpectations(t)
}

func TestWeatherService_GetWeatherByCity_EmptyName(t *testing.T) {
	provider := new(MockWeatherProvider)
	weatherService := service.NewWeatherService(provider)

	res, err := weatherService.GetWeatherByCity(context.Background(), "")

	require.ErrorIs(t, err, service.ErrCityNotFound)
	assert.Nil(t, res)
}

func TestWeatherService_GetWeatherByCity_NotFound(t *testing.T) {
	provider := new(MockWeatherProvider)
	weatherService := service.NewWeatherService(provider)

	provider.On("GetCityCoordinates", mock.Anything, "Atlantis", "").Return(nil, nil).Once()

	res, err := weatherService.GetWeatherByCity(context.Background(), "Atlantis")

	require.ErrorIs(t, err, service.ErrCityNotFound)
	assert.Nil(t, res)
	provider.AssertExpectations(t)
}

func TestWeatherService_GetWeatherByCountry_Success(t *testing.T) {
	provider := new(MockWeatherProvider)
	weatherService := service.NewWeatherService(provider)

	cities := []service.ProviderCity{
		{Name: "Almaty", Country: "Kazakhstan", Latitude: 43.2, Longitude: 76.8},
		{Name: "Astana", Country: "Kazakhstan", Latitude: 51.1, Longitude: 71.4},
	}

	provider.On("GetCitiesByCountry", mock.Anything, "Kazakhstan", 10).Return(cities, nil).Once()
	provider.On("GetCurrentWeather", mock.Anything, 43.2, 76.8).Return(&service.ProviderWeatherResponse{Temperature: 15.0}, nil).Once()
	provider.On("GetCurrentWeather", mock.Anything, 51.1, 71.4).Return(&service.ProviderWeatherResponse{Temperature: 5.0}, nil).Once()

	res, err := weatherService.GetWeatherByCountry(context.Background(), "Kazakhstan")

	require.NoError(t, err)
	assert.Equal(t, "Kazakhstan", res.Country)
	assert.Len(t, res.Cities, 2)
	assert.Equal(t, "Almaty", res.Cities[0].City)
	assert.Equal(t, 15.0, res.Cities[0].Temperature)
	assert.Equal(t, "Astana", res.Cities[1].City)
	assert.Equal(t, 5.0, res.Cities[1].Temperature)
}

func TestWeatherService_GetWeatherByCountry_EmptyCountry(t *testing.T) {
	provider := new(MockWeatherProvider)
	weatherService := service.NewWeatherService(provider)

	res, err := weatherService.GetWeatherByCountry(context.Background(), "")

	require.ErrorIs(t, err, service.ErrCountryNotFound)
	assert.Nil(t, res)
}

func TestWeatherService_GetWeatherByCountry_NotFound(t *testing.T) {
	provider := new(MockWeatherProvider)
	weatherService := service.NewWeatherService(provider)

	provider.On("GetCitiesByCountry", mock.Anything, "UnknownCountry", 10).Return(nil, nil).Once()

	res, err := weatherService.GetWeatherByCountry(context.Background(), "UnknownCountry")

	require.ErrorIs(t, err, service.ErrCountryNotFound)
	assert.Nil(t, res)
}

func TestWeatherService_GetTopWarmestCitiesByCountry_Success(t *testing.T) {
	provider := new(MockWeatherProvider)
	weatherService := service.NewWeatherService(provider)

	cities := []service.ProviderCity{
		{Name: "CityA", Country: "Kazakhstan", Latitude: 1.0, Longitude: 1.0},
		{Name: "CityB", Country: "Kazakhstan", Latitude: 2.0, Longitude: 2.0},
		{Name: "CityC", Country: "Kazakhstan", Latitude: 3.0, Longitude: 3.0},
		{Name: "CityD", Country: "Kazakhstan", Latitude: 4.0, Longitude: 4.0},
	}

	provider.On("GetCitiesByCountry", mock.Anything, "Kazakhstan", 10).Return(cities, nil).Once()
	provider.On("GetCurrentWeather", mock.Anything, 1.0, 1.0).Return(&service.ProviderWeatherResponse{Temperature: 10.0}, nil).Once()
	provider.On("GetCurrentWeather", mock.Anything, 2.0, 2.0).Return(&service.ProviderWeatherResponse{Temperature: 25.0}, nil).Once()
	provider.On("GetCurrentWeather", mock.Anything, 3.0, 3.0).Return(&service.ProviderWeatherResponse{Temperature: 15.0}, nil).Once()
	provider.On("GetCurrentWeather", mock.Anything, 4.0, 4.0).Return(&service.ProviderWeatherResponse{Temperature: 5.0}, nil).Once()

	res, err := weatherService.GetTopWarmestCitiesByCountry(context.Background(), "Kazakhstan")

	require.NoError(t, err)
	require.Len(t, res.Cities, 3)
	// Sorted descending: CityB (25), CityC (15), CityA (10)
	assert.Equal(t, "CityB", res.Cities[0].City)
	assert.Equal(t, 25.0, res.Cities[0].Temperature)
	assert.Equal(t, "CityC", res.Cities[1].City)
	assert.Equal(t, 15.0, res.Cities[1].Temperature)
	assert.Equal(t, "CityA", res.Cities[2].City)
	assert.Equal(t, 10.0, res.Cities[2].Temperature)
}

func TestWeatherService_MapWeatherCode_EdgeCases(t *testing.T) {
	// Let's test custom weather codes mapping to trigger mapWeatherCode and outfitByTemperature branches
	provider := new(MockWeatherProvider)
	weatherService := service.NewWeatherService(provider)

	// We can test how codes are mapped. Let's do GetWeather for codes: 45 (Туман), 51 (Морось), 61 (Дождь), 71 (Снег), 95 (Гроза), 999 (Неизвестно)
	codes := []int{45, 51, 61, 71, 95, 999}
	expectedDesc := []string{"Туман", "Морось", "Дождь", "Снег", "Гроза", "Неизвестно"}

	for i, code := range codes {
		provider.On("GetCurrentWeather", mock.Anything, 0.0, 0.0).Return(&service.ProviderWeatherResponse{
			Temperature: 5.0, // < 10 -> Тёплая одежда
			WeatherCode: code,
		}, nil).Once()

		res, err := weatherService.GetWeather(context.Background(), 0.0, 0.0)
		require.NoError(t, err)
		assert.Equal(t, expectedDesc[i], res.Description)
		assert.Equal(t, "Тёплая одежда", res.Outfit)
	}
}
