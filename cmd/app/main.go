package main

import (
	"log"
	"net/http"
	"time"
	"weather-api/internal/client"
	"weather-api/internal/handler"
	"weather-api/internal/service"

	"github.com/go-chi/chi/v5"
)

func main() {
	router := chi.NewRouter()

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	weatherClient := client.NewWeatherClient(httpClient)
	weatherService := service.NewWeatherService(weatherClient)
	weatherHandler := handler.NewWeatherHandler(weatherService)

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	router.Route("/api", func(r chi.Router) {
		r.Get("/weather", weatherHandler.GetWeather)
	})

	router.Get("/weather/{city}", weatherHandler.GetWeatherByCity)
	router.Get("/weather/country/{country}", weatherHandler.GetWeatherByCountry)
	router.Get("/weather/country/{country}/top", weatherHandler.GetTopWarmestCitiesByCountry)

	addr := ":8080"
	log.Printf("server started on %s", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal(err)
	}
}
