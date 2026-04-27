package service

import (
	"context"
	"sync"
	"weather-api/internal/domain"
)

type WeatherHistoryRepository interface {
	Save(ctx context.Context, input *domain.SaveWeatherHistoryInput) (domain.WeatherHistory, error)
	GetHistory(ctx context.Context, userID int64, filter domain.WeatherHistoryFilter) ([]domain.WeatherHistory, error)
}

type UserWeatherService struct {
	userService     *UserService
	userCityService *UserCityService
	weatherService  *WeatherService
	historyRepo     WeatherHistoryRepository
}

func NewUserWeatherService(
	userService *UserService,
	userCityService *UserCityService,
	weatherService *WeatherService,
	historyRepo WeatherHistoryRepository,
) *UserWeatherService {
	return &UserWeatherService{
		userService:     userService,
		userCityService: userCityService,
		weatherService:  weatherService,
		historyRepo:     historyRepo,
	}
}

// UserWeatherResult — структура ответа для GET /users/{id}/weather
type UserWeatherResult struct {
	UserID int64         `json:"user_id"`
	Cities []CityWeather `json:"cities"`
}

// GetUserWeather параллельно запрашивает погоду по всем городам юзера и сохраняет историю
func (s *UserWeatherService) GetUserWeather(ctx context.Context, userID int64) (*UserWeatherResult, error) {

	_, err := s.userService.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	cities, err := s.userCityService.ListCities(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(cities) == 0 {
		return &UserWeatherResult{UserID: userID, Cities: make([]CityWeather, 0)}, nil
	}

	// 3. Параллельный запрос к Weather API (Дополнительное задание выполнено)
	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make([]CityWeather, 0, len(cities))

	for _, userCity := range cities {
		wg.Add(1)
		go func(city domain.UserCity) {
			defer wg.Done()

			weather, err := s.weatherService.GetWeatherByCity(ctx, city.City)
			if err != nil {
				return
			}

			mu.Lock()
			results = append(results, *weather)
			mu.Unlock()

			_, _ = s.historyRepo.Save(ctx, &domain.SaveWeatherHistoryInput{
				UserID:      userID,
				City:        weather.City,
				Temperature: weather.Temperature,
				Description: weather.Description,
			})
		}(userCity)
	}

	wg.Wait()

	return &UserWeatherResult{
		UserID: userID,
		Cities: results,
	}, nil
}

// HistoryResponse — структура ответа для истории погоды (из ТЗ)
type HistoryResponse struct {
	UserID  int64                   `json:"user_id"`
	City    string                  `json:"city,omitempty"`
	History []domain.WeatherHistory `json:"history"`
}

// GetHistory возвращает историю с фильтрацией (город, лимит, оффсет)
func (s *UserWeatherService) GetHistory(ctx context.Context, userID int64, filter domain.WeatherHistoryFilter) (*HistoryResponse, error) {

	_, err := s.userService.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	history, err := s.historyRepo.GetHistory(ctx, userID, filter)
	if err != nil {
		return nil, err
	}

	return &HistoryResponse{
		UserID:  userID,
		City:    filter.City,
		History: history,
	}, nil
}
