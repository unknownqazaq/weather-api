package handler

import (
	"net/http"
	"weather-api/internal/auth"
	"weather-api/internal/domain"

	"github.com/go-chi/chi/v5"
)

func NewRouter(
	weatherHandler *WeatherHandler,
	userHandler *UserHandler,
	userCityHandler *UserCityHandler,
	userWeatherHandler *UserWeatherHandler,
	authHandler *AuthHandler,
	jwtManager *auth.JWTManager,
) *chi.Mux {
	router := chi.NewRouter()

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	router.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
	})

	router.Route("/api/v1", func(r chi.Router) {
		r.Use(auth.AuthMiddleware(jwtManager))

		r.Get("/users/me", userHandler.GetMe)

		// Admin only routes
		r.Group(func(r chi.Router) {
			r.Use(auth.RequireRole(domain.RoleAdmin))
			r.Get("/users", userHandler.List)
			r.Get("/users/{id}", userHandler.GetByID)
			r.Delete("/users/{id}", userHandler.Delete)
			r.Put("/users/{id}", userHandler.Update) // Also make admin only for simplicity unless specified
		})

		r.Post("/cities", userCityHandler.AddCity)
		r.Get("/cities", userCityHandler.ListCities)
		r.Delete("/cities/{city_id}", userCityHandler.DeleteCity)

		r.Get("/weather", userWeatherHandler.GetWeather)
		r.Get("/weather/history", userWeatherHandler.GetHistory)
	})

	// Public weather endpoints (if any)
	router.Route("/api", func(r chi.Router) {
		r.Get("/weather_coords", weatherHandler.GetWeather) // Renamed to avoid collision with protected /weather
	})
	router.Get("/weather/city/{city}", weatherHandler.GetWeatherByCity)
	router.Get("/weather/country/{country}", weatherHandler.GetWeatherByCountry)
	router.Get("/weather/country/{country}/top", weatherHandler.GetTopWarmestCitiesByCountry)

	return router
}
