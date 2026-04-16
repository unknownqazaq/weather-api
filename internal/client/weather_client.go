package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

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
		geocodingURL:     "https://geocoding-api.open-meteo.com/v1/search",
		countryCitiesURL: "https://countriesnow.space/api/v0.1/countries/cities",
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

type geocodingResponse struct {
	Results []struct {
		Name       string  `json:"name"`
		Country    string  `json:"country"`
		Latitude   float64 `json:"latitude"`
		Longitude  float64 `json:"longitude"`
		Population int     `json:"population"`
	}
}

type countriesNowResponse struct {
	Error bool     `json:"error"`
	Msg   string   `json:"msg"`
	Data  []string `json:"data"`
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

func (c *WeatherClient) GetCityCoordinates(ctx context.Context, city, country string) (*service.ProviderCity, error) {
	u, err := url.Parse(c.geocodingURL)
	if err != nil {
		return nil, fmt.Errorf("parse geocoding url: %w", err)
	}

	q := u.Query()
	q.Set("name", city)
	q.Set("count", "10")
	q.Set("language", "en")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create geocoding request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call geocoding api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geocoding api returned status: %d", resp.StatusCode)
	}

	var result geocodingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode geocoding response: %w", err)
	}

	if len(result.Results) == 0 {
		return nil, nil
	}

	var selected *service.ProviderCity
	for _, item := range result.Results {
		if country != "" && !strings.EqualFold(item.Country, country) {
			continue
		}

		candidate := &service.ProviderCity{
			Name:       item.Name,
			Country:    item.Country,
			Latitude:   item.Latitude,
			Longitude:  item.Longitude,
			Population: item.Population,
		}

		if selected == nil || candidate.Population > selected.Population {
			selected = candidate
		}
	}

	if selected == nil {
		return nil, nil
	}

	return selected, nil
}

func (c *WeatherClient) GetCitiesByCountry(ctx context.Context, country string, limit int) ([]service.ProviderCity, error) {
	if limit <= 0 {
		limit = 10
	}

	payload, err := json.Marshal(map[string]string{"country": country})
	if err != nil {
		return nil, fmt.Errorf("marshal country request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.countryCitiesURL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("create country cities request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call country cities api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("country cities api returned status: %d", resp.StatusCode)
	}

	var countriesResp countriesNowResponse
	if err := json.NewDecoder(resp.Body).Decode(&countriesResp); err != nil {
		return nil, fmt.Errorf("decode country cities response: %w", err)
	}

	if countriesResp.Error || len(countriesResp.Data) == 0 {
		return nil, nil
	}

	cities := make([]service.ProviderCity, 0, limit)
	seen := make(map[string]struct{})
	var lastErr error

	for _, cityName := range countriesResp.Data {
		if len(cities) >= limit {
			break
		}

		city, err := c.GetCityCoordinates(ctx, cityName, country)
		if err != nil {
			lastErr = err
			continue
		}
		if city == nil {
			continue
		}

		key := strings.ToLower(strings.TrimSpace(city.Name))
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		cities = append(cities, *city)
	}

	if len(cities) == 0 && lastErr != nil {
		return nil, fmt.Errorf("geocode country cities: %w", lastErr)
	}

	return cities, nil
}
