package main

import (
	"log"
	"net/http"
	"time"
	"weather-api/internal/client"
	"weather-api/internal/config"
	"weather-api/internal/handler"
	"weather-api/internal/repository/postgres"
	"weather-api/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.MustLoad()

	db, err := sqlx.Connect("postgres", cfg.Database.DSN())
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()

	router := chi.NewRouter()

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	weatherClient := client.NewWeatherClient(httpClient)
	weatherService := service.NewWeatherService(weatherClient)
	weatherHandler := handler.NewWeatherHandler(weatherService)

	userRepo := postgres.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	router.Route("/api/v1/users", func(r chi.Router) {
		r.Post("/", userHandler.Create)
		r.Get("/{id}", userHandler.GetByID)
	})

	router.Route("/api", func(r chi.Router) {
		r.Get("/weather", weatherHandler.GetWeather)
	})
	router.Get("/weather/{city}", weatherHandler.GetWeatherByCity)
	router.Get("/weather/country/{country}", weatherHandler.GetWeatherByCountry)
	router.Get("/weather/country/{country}/top", weatherHandler.GetTopWarmestCitiesByCountry)

	addr := ":" + cfg.App.Port
	log.Printf("server started on %s", addr)

	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  cfg.App.ReadTimeout,
		WriteTimeout: cfg.App.WriteTimeout,
		IdleTimeout:  cfg.App.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
