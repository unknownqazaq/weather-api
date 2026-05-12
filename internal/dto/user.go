package dto

import (
	"errors"
	"strings"
	"time"
	"weather-api/internal/model"
)

var ErrInvalidUserInput = errors.New("invalid user input")

type CreateUserRequest struct {
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Role         string `json:"role"`
}

func (in *CreateUserRequest) NormalizeAndValidate() error {
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	in.PasswordHash = strings.TrimSpace(in.PasswordHash)
	in.FirstName = strings.TrimSpace(in.FirstName)
	in.LastName = strings.TrimSpace(in.LastName)
	in.Role = strings.TrimSpace(in.Role)
	if in.Role == "" {
		in.Role = model.RoleUser
	}

	if in.Email == "" || !strings.Contains(in.Email, "@") {
		return ErrInvalidUserInput
	}
	if in.PasswordHash == "" || in.FirstName == "" || in.LastName == "" {
		return ErrInvalidUserInput
	}
	return nil
}

type UpdateUserRequest struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
}

func (in *UpdateUserRequest) Validate() error {
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

type UserResponse struct {
	ID        int64      `json:"id"`
	Email     string     `json:"email"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Role      string     `json:"role"`
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

func MapUser(user model.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		DeletedAt: user.DeletedAt,
	}
}
