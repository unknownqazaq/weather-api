package postgres

import (
	"context"
	"weather-api/internal/domain"

	"github.com/jmoiron/sqlx"
)

type UserCityRepository struct {
	db *sqlx.DB
}

func NewUserCityRepository(db *sqlx.DB) *UserCityRepository {
	return &UserCityRepository{db: db}
}

func (r *UserCityRepository) AddCity(ctx context.Context, input *domain.AddUserCityInput) (domain.UserCity, error) {
	query := `
		INSERT INTO user_cities (user_id, city)
		VALUES (:user_id, :city)
		RETURNING *
	`

	var userCity domain.UserCity
	rows, err := r.db.NamedQueryContext(ctx, query, input)
	if err != nil {
		return domain.UserCity{}, err
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.StructScan(&userCity); err != nil {
			return domain.UserCity{}, err
		}
	}

	return userCity, nil
}

func (r *UserCityRepository) ListCities(ctx context.Context, userID int64) ([]domain.UserCity, error) {
	query := `SELECT * FROM user_cities WHERE user_id = $1 ORDER BY added_at DESC`

	var cities []domain.UserCity
	err := r.db.SelectContext(ctx, &cities, query, userID)
	if err != nil {
		return nil, err
	}

	if cities == nil {
		cities = make([]domain.UserCity, 0)
	}

	return cities, nil
}

func (r *UserCityRepository) DeleteCity(ctx context.Context, userID int64, cityID int64) error {
	query := `DELETE FROM user_cities WHERE id = $1 AND user_id = $2`
	_, err := r.db.ExecContext(ctx, query, cityID, userID)
	return err
}
