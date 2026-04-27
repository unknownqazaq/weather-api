package postgres

import (
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v5/stdlib" // pgx driver
	"weather-api/internal/config"
)

// NewDB открывает соединение с базой данных и настраивает пул соединений
func NewDB(cfg config.DatabaseConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect("pgx", cfg.DSN())
	if err != nil {
		return nil, err
	}

	// Настройки пула соединений
	db.SetMaxOpenConns(10)                  // Максимум открытых соединений
	db.SetMaxIdleConns(5)                   // Максимум простаивающих соединений
	db.SetConnMaxLifetime(30 * time.Minute) // Время жизни соединения
	db.SetConnMaxIdleTime(5 * time.Minute)  // Время жизни простаивающего соединения

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
