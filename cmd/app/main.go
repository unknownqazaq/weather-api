package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"
	"weather-api/internal/auth"
	"weather-api/internal/client"
	"weather-api/internal/config"
	"weather-api/internal/handler"
	"weather-api/internal/repository/postgres"
	"weather-api/internal/service"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg := config.MustLoad()

	db, err := postgres.NewDB(cfg.Database)
	if err != nil {
		logger.Fatal("failed to connect to db", zap.Error(err))
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

	jwtManager := auth.NewJWTManager(cfg.App.JWTSecret, cfg.App.JWTExpiration)
	authService := service.NewAuthService(userService, userRepo, jwtManager)
	authHandler := handler.NewAuthHandler(authService)

	router := handler.NewRouter(
		weatherHandler,
		userHandler,
		userCityHandler,
		userWeatherHandler,
		authHandler,
		jwtManager,
		logger,
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
		logger.Info("server started", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("listen server", zap.Error(err))
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown server error", zap.Error(err))
	}

	logger.Info("server stopped gracefully")
}
