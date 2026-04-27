package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidUserID     = errors.New("invalid user id")
	ErrInvalidUserInput  = errors.New("invalid user input")
	ErrEmailAlreadyTaken = errors.New("email already exists")
)

type User struct {
	ID           int64      `db:"id" json:"id"`
	Email        string     `db:"email" json:"email"`
	PasswordHash string     `db:"password_hash" json:"-"`
	FirstName    string     `db:"first_name" json:"first_name"`
	LastName     string     `db:"last_name" json:"last_name"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	DeletedAt    *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

type CreateUserInput struct {
	Email        string `db:"email" json:"email"`
	PasswordHash string `db:"password_hash" json:"password_hash"`
	FirstName    string `db:"first_name" json:"first_name"`
	LastName     string `db:"last_name" json:"last_name"`
}

func (in *CreateUserInput) NormalizeAndValidate() error {
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	in.PasswordHash = strings.TrimSpace(in.PasswordHash)
	in.FirstName = strings.TrimSpace(in.FirstName)
	in.LastName = strings.TrimSpace(in.LastName)

	if in.Email == "" || !strings.Contains(in.Email, "@") {
		return ErrInvalidUserInput
	}
	if in.PasswordHash == "" || in.FirstName == "" || in.LastName == "" {
		return ErrInvalidUserInput
	}
	return nil
}

type UpdateUserInput struct {
	FirstName *string `db:"first_name" json:"first_name"`
	LastName  *string `db:"last_name" json:"last_name"`
}

func (in *UpdateUserInput) Validate() error {
	if in.FirstName != nil {
		*in.FirstName = strings.TrimSpace(*in.FirstName)
		if *in.FirstName == "" {
			return ErrInvalidUserInput
		}
	}
	if in.LastName != nil {
		*in.LastName = strings.TrimSpace(*in.LastName)
		if *in.LastName == "" {
			return ErrInvalidUserInput
		}
	}
	return nil
}

type ListUsersFilter struct {
	Limit          int
	Offset         int
	Query          string
	IncludeDeleted bool
}

func (f *ListUsersFilter) Normalize() {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	f.Query = strings.TrimSpace(strings.ToLower(f.Query))
}
