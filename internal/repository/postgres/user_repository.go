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
func (r *userRepository) List(ctx context.Context, filter domain.ListUsersFilter) ([]domain.User, error) {
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
		args["query"] = "%" + strings.ToLower(filter.Query) + "%"
	}

	builder.WriteString(" ORDER BY created_at DESC LIMIT :limit OFFSET :offset")

	query, queryArgs, err := sqlx.Named(builder.String(), args)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)

	var users []domain.User
	err = r.db.SelectContext(ctx, &users, query, queryArgs...)
	if err != nil {
		return nil, err
	}
	if users == nil {
		users = make([]domain.User, 0)
	}
	return users, nil
}

func (r *userRepository) Update(ctx context.Context, id int64, input *domain.UpdateUserInput) (domain.User, error) {
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
		return domain.User{}, err
	}
	query = r.db.Rebind(query)

	var user domain.User
	err = r.db.QueryRowxContext(ctx, query, queryArgs...).StructScan(&user)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, err
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
		return domain.ErrUserNotFound
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
