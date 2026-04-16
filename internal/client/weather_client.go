package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"weather-api/internal/service"
)

type WeatherClient struct {
	httpClient       *http.Client
	forecastURL      string
	geocodingURL     string
	countryCitiesURL string
}

func NewWeatherClient(httpClient *http.Client) *WeatherClient {
	return &WeatherClient{
		httpClient:       httpClient,
		forecastURL:      "https://api.open-meteo.com/v1/forecast",
	}
}

type openMeteoResponse struct {
	CurrentWeather struct {
		Temperature float64 `json:"temperature"`
		Windspeed   float64 `json:"windspeed"`
		Weathercode int     `json:"weathercode"`
		Time        string  `json:"time"`
	} `json:"current_weather"`
}



func (c *WeatherClient) GetCurrentWeather(ctx context.Context, lat, lon float64) (*service.ProviderWeatherResponse, error) {
	u, err := url.Parse(c.forecastURL)
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}

	q := u.Query()
	q.Set("latitude", fmt.Sprintf("%.4f", lat))
	q.Set("longitude", fmt.Sprintf("%.4f", lon))
	q.Set("current_weather", "true")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call external api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("external api returned status: %d", resp.StatusCode)
	}

	var result openMeteoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode external api response: %w", err)
	}

	return &service.ProviderWeatherResponse{
		Temperature: result.CurrentWeather.Temperature,
		WindSpeed:   result.CurrentWeather.Windspeed,
		WeatherCode: result.CurrentWeather.Weathercode,
		Time:        result.CurrentWeather.Time,
	}, nil
}
