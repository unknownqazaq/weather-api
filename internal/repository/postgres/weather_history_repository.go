package postgres

import (
	"context"
	"strings"
	"weather-api/internal/domain"

	"github.com/jmoiron/sqlx"
)

type WeatherHistoryRepository struct {
	db *sqlx.DB
}

func NewWeatherHistoryRepository(db *sqlx.DB) *WeatherHistoryRepository {
	return &WeatherHistoryRepository{db: db}
}

func (r *WeatherHistoryRepository) Save(ctx context.Context, input *domain.SaveWeatherHistoryInput) (domain.WeatherHistory, error) {
	query := `
		INSERT INTO weather_history (user_id, city, temperature, description)
		VALUES (:user_id, :city, :temperature, :description)
		RETURNING *
	`

	var history domain.WeatherHistory
	rows, err := r.db.NamedQueryContext(ctx, query, input)
	if err != nil {
		return domain.WeatherHistory{}, err
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.StructScan(&history); err != nil {
			return domain.WeatherHistory{}, err
		}
	}

	return history, nil
}

func (r *WeatherHistoryRepository) GetHistory(ctx context.Context, userID int64, filter domain.WeatherHistoryFilter) ([]domain.WeatherHistory, error) {
	var builder strings.Builder
	builder.WriteString("SELECT * FROM weather_history WHERE user_id = :user_id")

	args := map[string]interface{}{
		"user_id": userID,
	}

	if filter.City != "" {
		builder.WriteString(" AND city = :city")
		args["city"] = filter.City
	}

	builder.WriteString(" ORDER BY requested_at DESC")

	if filter.Limit > 0 {
		builder.WriteString(" LIMIT :limit")
		args["limit"] = filter.Limit
	}

	if filter.Offset > 0 {
		builder.WriteString(" OFFSET :offset")
		args["offset"] = filter.Offset
	}

	query, queryArgs, err := sqlx.Named(builder.String(), args)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)

	var history []domain.WeatherHistory
	err = r.db.SelectContext(ctx, &history, query, queryArgs...)
	if err != nil {
		return nil, err
	}

	if history == nil {
		history = make([]domain.WeatherHistory, 0)
	}

	return history, nil
}
