package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"weather-api/internal/service"
)

type WeatherClient struct {
	httpClient *http.Client
	gatewayURL string
}

func NewWeatherClient(httpClient *http.Client, gatewayURL string) *WeatherClient {
	return &WeatherClient{
		httpClient: httpClient,
		gatewayURL: strings.TrimSuffix(gatewayURL, "/"),
	}
}

func (c *WeatherClient) GetCurrentWeather(ctx context.Context, lat, lon float64) (*service.ProviderWeatherResponse, error) {
	u, err := url.Parse(fmt.Sprintf("%s/external/weather", c.gatewayURL))
	if err != nil {
		return nil, fmt.Errorf("parse gateway url: %w", err)
	}

	q := u.Query()
	q.Set("latitude", fmt.Sprintf("%.4f", lat)) // Wait, the gateway endpoint /external/weather expects lat and lon query parameters!
	// Let's check what gateway-service/cmd/main.go expects:
	// latStr := r.URL.Query().Get("lat")
	// lonStr := r.URL.Query().Get("lon")
	// Ah! The gateway-service expects "lat" and "lon"!
	q.Set("lat", fmt.Sprintf("%.4f", lat))
	q.Set("lon", fmt.Sprintf("%.4f", lon))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call gateway-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error string `json:"error"`
		}
		if json.NewDecoder(resp.Body).Decode(&errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("gateway-service returned status %d: %s", resp.StatusCode, errResp.Error)
		}
		return nil, fmt.Errorf("gateway-service returned status: %d", resp.StatusCode)
	}

	// Define response structure matching gateway-service's client.ProviderWeatherResponse JSON keys
	var result struct {
		Temperature float64 `json:"temperature"`
		WindSpeed   float64 `json:"wind_speed"`
		WeatherCode int     `json:"weather_code"`
		Time        string  `json:"time"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode gateway-service response: %w", err)
	}

	return &service.ProviderWeatherResponse{
		Temperature: result.Temperature,
		WindSpeed:   result.WindSpeed,
		WeatherCode: result.WeatherCode,
		Time:        result.Time,
	}, nil
}

func (c *WeatherClient) GetCityCoordinates(ctx context.Context, city, country string) (*service.ProviderCity, error) {
	u, err := url.Parse(fmt.Sprintf("%s/external/geocode", c.gatewayURL))
	if err != nil {
		return nil, fmt.Errorf("parse gateway url: %w", err)
	}

	q := u.Query()
	q.Set("city", city)
	if country != "" {
		q.Set("country", country)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call gateway-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error string `json:"error"`
		}
		if json.NewDecoder(resp.Body).Decode(&errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("gateway-service returned status %d: %s", resp.StatusCode, errResp.Error)
		}
		return nil, fmt.Errorf("gateway-service returned status: %d", resp.StatusCode)
	}

	// Define response structure matching gateway-service's client.ProviderCity JSON keys
	var result struct {
		Name       string  `json:"name"`
		Country    string  `json:"country"`
		Latitude   float64 `json:"latitude"`
		Longitude  float64 `json:"longitude"`
		Population int     `json:"population"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode gateway-service response: %w", err)
	}

	return &service.ProviderCity{
		Name:       result.Name,
		Country:    result.Country,
		Latitude:   result.Latitude,
		Longitude:  result.Longitude,
		Population: result.Population,
	}, nil
}

func (c *WeatherClient) GetCitiesByCountry(ctx context.Context, country string, limit int) ([]service.ProviderCity, error) {
	u, err := url.Parse(fmt.Sprintf("%s/external/cities", c.gatewayURL))
	if err != nil {
		return nil, fmt.Errorf("parse gateway url: %w", err)
	}

	q := u.Query()
	q.Set("country", country)
	q.Set("limit", fmt.Sprintf("%d", limit))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call gateway-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error string `json:"error"`
		}
		if json.NewDecoder(resp.Body).Decode(&errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("gateway-service returned status %d: %s", resp.StatusCode, errResp.Error)
		}
		return nil, fmt.Errorf("gateway-service returned status: %d", resp.StatusCode)
	}

	// Define response structure matching gateway-service's []client.ProviderCity JSON keys
	var result []struct {
		Name       string  `json:"name"`
		Country    string  `json:"country"`
		Latitude   float64 `json:"latitude"`
		Longitude  float64 `json:"longitude"`
		Population int     `json:"population"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode gateway-service response: %w", err)
	}

	cities := make([]service.ProviderCity, len(result))
	for i, item := range result {
		cities[i] = service.ProviderCity{
			Name:       item.Name,
			Country:    item.Country,
			Latitude:   item.Latitude,
			Longitude:  item.Longitude,
			Population: item.Population,
		}
	}

	return cities, nil
}
