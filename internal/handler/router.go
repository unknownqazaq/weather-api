package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter(
	weatherHandler *WeatherHandler,
	userHandler *UserHandler,
	userCityHandler *UserCityHandler,
	userWeatherHandler *UserWeatherHandler,
) *chi.Mux {
	router := chi.NewRouter()

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	router.Route("/api/v1/users", func(r chi.Router) {
		r.Post("/", userHandler.Create)
		r.Get("/", userHandler.List)
		r.Get("/{id}", userHandler.GetByID)
		r.Put("/{id}", userHandler.Update)
		r.Delete("/{id}", userHandler.Delete)

		r.Post("/{id}/cities", userCityHandler.AddCity)
		r.Get("/{id}/cities", userCityHandler.ListCities)
		r.Delete("/{id}/cities/{city_id}", userCityHandler.DeleteCity)

		r.Get("/{id}/weather", userWeatherHandler.GetWeather)
		r.Get("/{id}/weather/history", userWeatherHandler.GetHistory)
	})

	router.Route("/api", func(r chi.Router) {
		r.Get("/weather", weatherHandler.GetWeather)
	})
	router.Get("/weather/{city}", weatherHandler.GetWeatherByCity)
	router.Get("/weather/country/{country}", weatherHandler.GetWeatherByCountry)
	router.Get("/weather/country/{country}/top", weatherHandler.GetTopWarmestCitiesByCountry)

	return router
}
