package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
	"weather-api/internal/client"
	"weather-api/internal/config"
	"weather-api/internal/handler"
	"weather-api/internal/repository/postgres"
	"weather-api/internal/service"

	_ "github.com/lib/pq"
)

func main() {
	cfg := config.MustLoad()

	db, err := postgres.NewDB(cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	weatherClient := client.NewWeatherClient(httpClient)
	weatherService := service.NewWeatherService(weatherClient)
	weatherHandler := handler.NewWeatherHandler(weatherService)

	userRepo := postgres.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	userCityRepo := postgres.NewUserCityRepository(db)
	userCityService := service.NewUserCityService(userCityRepo, userService)
	userCityHandler := handler.NewUserCityHandler(userCityService)

	weatherHistoryRepo := postgres.NewWeatherHistoryRepository(db)
	userWeatherService := service.NewUserWeatherService(userService, userCityService, weatherService, weatherHistoryRepo)
	userWeatherHandler := handler.NewUserWeatherHandler(userWeatherService)

	router := handler.NewRouter(
		weatherHandler,
		userHandler,
		userCityHandler,
		userWeatherHandler,
	)

	addr := ":" + cfg.App.Port

	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  cfg.App.ReadTimeout,
		WriteTimeout: cfg.App.WriteTimeout,
		IdleTimeout:  cfg.App.IdleTimeout,
	}

	go func() {
		log.Printf("server started on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen server: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown server error: %v", err)
	}

	log.Println("server stopped gracefully")
}
