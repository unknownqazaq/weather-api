package config

import (
	"fmt"
	"log"
	"os"
	"time"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
}

type AppConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	SSLMode  string
}

func MustLoad() Config {
	cfg := Config{
		App: AppConfig{
			Port:         getEnv("APP_PORT", "8080"),
			ReadTimeout:  mustDuration("APP_READ_TIMEOUT", "5s"),
			WriteTimeout: mustDuration("APP_WRITE_TIMEOUT", "10s"),
			IdleTimeout:  mustDuration("APP_IDLE_TIMEOUT", "60s"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "127.0.0.1"),
			Port:     getEnv("DB_PORT", "5433"), // Порт изменен на 5433
			Name:     getEnv("DB_NAME", "users_db"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
	}

	return cfg
}

func (c DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		c.Host,
		c.Port,
		c.Name,
		c.User,
		c.Password,
		c.SSLMode,
	)
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func mustDuration(key, fallback string) time.Duration {
	value := getEnv(key, fallback)
	d, err := time.ParseDuration(value)
	if err != nil {
		log.Fatalf("invalid duration %s=%s: %v", key, value, err)
	}
	return d
}
