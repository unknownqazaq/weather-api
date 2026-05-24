package postgres

import (
	"context"
	"errors"
	"strings"
	"weather-api/internal/dto"
	"weather-api/internal/model"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *userRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) (model.User, error) {

	query := `INSERT INTO users(email, password_hash, first_name, last_name, role)
			VALUES(:email, :password_hash, :first_name, :last_name, :role)
			RETURNING *`

	rows, err := r.db.NamedQueryContext(ctx, query, user)
	if err != nil {
		if isUniqueViolation(err) {
			return model.User{}, model.ErrEmailAlreadyTaken
		}
		return model.User{}, err
	}
	defer rows.Close()

	if rows.Next() {
		var user model.User
		if err := rows.StructScan(&user); err != nil {
			return model.User{}, err
		}
		return user, nil
	}

	return model.User{}, errors.New("failed to insert user")
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (model.User, error) {
	query := `SELECT * FROM users WHERE id=$1 AND deleted_at IS NULL`
	var user model.User
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return model.User{}, model.ErrUserNotFound
		}
		return model.User{}, err
	}
	return user, nil

}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (model.User, error) {
	query := `SELECT * FROM users WHERE email=$1 AND deleted_at IS NULL`
	var user model.User
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return model.User{}, model.ErrUserNotFound
		}
		return model.User{}, err
	}
	return user, nil
}

func (r *userRepository) List(ctx context.Context, filter dto.ListUsersFilter) ([]model.User, error) {
	var builder strings.Builder
	builder.WriteString("SELECT * FROM users WHERE 1=1")

	args := map[string]interface{}{
		"limit":  filter.Limit,
		"offset": filter.Offset,
	}

	if !filter.IncludeDeleted {
		builder.WriteString(" AND deleted_at IS NULL")
	}

	if filter.Query != "" {
		builder.WriteString(" AND (LOWER(email) LIKE :query OR LOWER(first_name) LIKE :query OR LOWER(last_name) LIKE :query)")
		args["query"] = "%" + filter.Query + "%"
	}

	builder.WriteString(" ORDER BY created_at DESC LIMIT :limit OFFSET :offset")

	query, queryArgs, err := sqlx.Named(builder.String(), args)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)

	var users []model.User
	err = r.db.SelectContext(ctx, &users, query, queryArgs...)
	if err != nil {
		return nil, err
	}
	if users == nil {
		users = make([]model.User, 0)
	}
	return users, nil
}

func (r *userRepository) Update(ctx context.Context, id int64, input *dto.UpdateUserRequest) (model.User, error) {
	var builder strings.Builder
	builder.WriteString("UPDATE users SET ")

	args := map[string]interface{}{
		"id": id,
	}
	var setClauses []string

	if input.FirstName != nil {
		setClauses = append(setClauses, "first_name = :first_name")
		args["first_name"] = *input.FirstName
	}
	if input.LastName != nil {
		setClauses = append(setClauses, "last_name = :last_name")
		args["last_name"] = *input.LastName
	}

	builder.WriteString(strings.Join(setClauses, ", "))
	builder.WriteString(" WHERE id = :id AND deleted_at IS NULL RETURNING *")

	query, queryArgs, err := sqlx.Named(builder.String(), args)
	if err != nil {
		return model.User{}, err
	}
	query = r.db.Rebind(query)

	var user model.User
	err = r.db.QueryRowxContext(ctx, query, queryArgs...).StructScan(&user)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return model.User{}, model.ErrUserNotFound
		}
		return model.User{}, err
	}
	return user, nil
}

func (r *userRepository) Delete(ctx context.Context, id int64) error {
	query := `UPDATE users SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return model.ErrUserNotFound
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return strings.Contains(strings.ToLower(err.Error()), "duplicate key")
}
