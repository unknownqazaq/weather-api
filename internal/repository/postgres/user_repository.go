package postgres

import (
	"context"
	"errors"
	"strings"
	"weather-api/internal/domain"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *userRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, input *domain.CreateUserInput) (domain.User, error) {

	query := `INSERT INTO users(email, password_hash, first_name, last_name)
			VALUES(:email, :password_hash, :first_name, :last_name)
			RETURNING *`

	rows, err := r.db.NamedQueryContext(ctx, query, input)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.User{}, domain.ErrEmailAlreadyTaken
		}
		return domain.User{}, err
	}
	defer rows.Close()

	if rows.Next() {
		var user domain.User
		if err := rows.StructScan(&user); err != nil {
			return domain.User{}, err
		}
		return user, nil
	}

	return domain.User{}, errors.New("failed to insert user")
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (domain.User, error) {
	query := `SELECT * FROM users WHERE id=$1 AND deleted_at IS NULL`
	var user domain.User
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, err
	}
	return user, nil

}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return strings.Contains(strings.ToLower(err.Error()), "duplicate key")
}
