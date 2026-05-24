package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
)

type WeatherProvider interface {
	GetCurrentWeather(ctx context.Context, lat, lon float64) (*ProviderWeatherResponse, error)
	GetCityCoordinates(ctx context.Context, city, country string) (*ProviderCity, error)
	GetCitiesByCountry(ctx context.Context, country string, limit int) ([]ProviderCity, error)
}

type ProviderWeatherResponse struct {
	Temperature float64
	WindSpeed   float64
	WeatherCode int
	Time        string
}

type ProviderCity struct {
	Name       string
	Country    string
	Latitude   float64
	Longitude  float64
	Population int
}

type WeatherResult struct {
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Temperature float64 `json:"temperature"`
	WindSpeed   float64 `json:"wind_speed"`
	WeatherCode int     `json:"weather_code"`
	Time        string  `json:"time"`
	Description string  `json:"description"`
	Outfit      string  `json:"outfit_recommendation"`
}

type CityWeather struct {
	City        string  `json:"city"`
	Country     string  `json:"country"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Temperature float64 `json:"temperature"`
	WindSpeed   float64 `json:"wind_speed"`
	WeatherCode int     `json:"weather_code"`
	Time        string  `json:"time"`
	Description string  `json:"description"`
	Outfit      string  `json:"outfit_recommendation"`
}

type CountryWeatherResult struct {
	Country string        `json:"country"`
	Cities  []CityWeather `json:"cities"`
}

type TopWarmestResult struct {
	Country string        `json:"country"`
	Cities  []CityWeather `json:"cities"`
}

var (
	ErrCityNotFound    = errors.New("city not found")
	ErrCountryNotFound = errors.New("country not found")
)

type WeatherService struct {
	provider WeatherProvider
}

func NewWeatherService(provider WeatherProvider) *WeatherService {
	return &WeatherService{
		provider: provider,
	}
}

func (s *WeatherService) GetWeather(ctx context.Context, lat, lon float64) (*WeatherResult, error) {
	resp, err := s.provider.GetCurrentWeather(ctx, lat, lon)
	if err != nil {
		return nil, fmt.Errorf("get weather from provider: %w", err)
	}

	return &WeatherResult{
		Latitude:    lat,
		Longitude:   lon,
		Temperature: resp.Temperature,
		WindSpeed:   resp.WindSpeed,
		WeatherCode: resp.WeatherCode,
		Time:        resp.Time,
		Description: mapWeatherCode(resp.WeatherCode),
		Outfit:      outfitByTemperature(resp.Temperature),
	}, nil
}

func (s *WeatherService) GetWeatherByCity(ctx context.Context, city string) (*CityWeather, error) {
	city = strings.TrimSpace(city)
	if city == "" {
		return nil, ErrCityNotFound
	}

	coords, err := s.provider.GetCityCoordinates(ctx, city, "")
	if err != nil {
		return nil, fmt.Errorf("get city coordinates: %w", err)
	}
	if coords == nil {
		return nil, ErrCityNotFound
	}

	weather, err := s.provider.GetCurrentWeather(ctx, coords.Latitude, coords.Longitude)
	if err != nil {
		return nil, fmt.Errorf("get weather by city: %w", err)
	}

	return &CityWeather{
		City:        coords.Name,
		Country:     coords.Country,
		Latitude:    coords.Latitude,
		Longitude:   coords.Longitude,
		Temperature: weather.Temperature,
		WindSpeed:   weather.WindSpeed,
		WeatherCode: weather.WeatherCode,
		Time:        weather.Time,
		Description: mapWeatherCode(weather.WeatherCode),
		Outfit:      outfitByTemperature(weather.Temperature),
	}, nil
}

func (s *WeatherService) GetWeatherByCountry(ctx context.Context, country string) (*CountryWeatherResult, error) {
	country = strings.TrimSpace(country)
	if country == "" {
		return nil, ErrCountryNotFound
	}

	cities, err := s.provider.GetCitiesByCountry(ctx, country, 10)
	if err != nil {
		return nil, fmt.Errorf("get cities by country: %w", err)
	}
	if len(cities) == 0 {
		return nil, ErrCountryNotFound
	}

	cityWeather := make([]CityWeather, 0, len(cities))
	for _, city := range cities {
		weather, err := s.provider.GetCurrentWeather(ctx, city.Latitude, city.Longitude)
		if err != nil {
			continue
		}

		cityWeather = append(cityWeather, CityWeather{
			City:        city.Name,
			Country:     city.Country,
			Latitude:    city.Latitude,
			Longitude:   city.Longitude,
			Temperature: weather.Temperature,
			WindSpeed:   weather.WindSpeed,
			WeatherCode: weather.WeatherCode,
			Time:        weather.Time,
			Description: mapWeatherCode(weather.WeatherCode),
			Outfit:      outfitByTemperature(weather.Temperature),
		})
	}

	if len(cityWeather) == 0 {
		return nil, fmt.Errorf("failed to load weather for country cities")
	}

	return &CountryWeatherResult{
		Country: country,
		Cities:  cityWeather,
	}, nil
}

func (s *WeatherService) GetTopWarmestCitiesByCountry(ctx context.Context, country string) (*TopWarmestResult, error) {
	result, err := s.GetWeatherByCountry(ctx, country)
	if err != nil {
		return nil, err
	}

	sort.Slice(result.Cities, func(i, j int) bool {
		return result.Cities[i].Temperature > result.Cities[j].Temperature
	})

	topCount := 3
	if len(result.Cities) < topCount {
		topCount = len(result.Cities)
	}

	return &TopWarmestResult{
		Country: result.Country,
		Cities:  result.Cities[:topCount],
	}, nil
}

func mapWeatherCode(code int) string {
	switch code {
	case 0:
		return "Ясно"
	case 1, 2, 3:
		return "Переменная облачность"
	case 45, 48:
		return "Туман"
	case 51, 53, 55:
		return "Морось"
	case 61, 63, 65:
		return "Дождь"
	case 71, 73, 75:
		return "Снег"
	case 95:
		return "Гроза"
	default:
		return "Неизвестно"
	}
}

func outfitByTemperature(temp float64) string {
	switch {
	case temp < 10:
		return "Тёплая одежда"
	case temp < 20:
		return "Куртка"
	default:
		return "Лёгкая одежда"
	}
}
