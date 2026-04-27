package service

import (
	"context"
	"weather-api/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

type userRepository interface {
	Create(ctx context.Context, input *domain.CreateUserInput) (domain.User, error)
	GetByID(ctx context.Context, id int64) (domain.User, error)
	List(ctx context.Context, filter domain.ListUsersFilter) ([]domain.User, error)
	Update(ctx context.Context, id int64, input *domain.UpdateUserInput) (domain.User, error)
	Delete(ctx context.Context, id int64) error
}

type UserService struct {
	repo userRepository
}

func NewUserService(repo userRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Create(ctx context.Context, input *domain.CreateUserInput) (domain.User, error) {
	if err := input.NormalizeAndValidate(); err != nil {
		return domain.User{}, err
	}
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(input.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, err
	}
	input.PasswordHash = string(hashBytes)

	return s.repo.Create(ctx, input)
}

func (s *UserService) GetByID(ctx context.Context, id int64) (domain.User, error) {
	if id <= 0 {
		return domain.User{}, domain.ErrInvalidUserID
	}
	return s.repo.GetByID(ctx, id)
}

func (s *UserService) List(ctx context.Context, filter domain.ListUsersFilter) ([]domain.User, error) {
	filter.Normalize()
	return s.repo.List(ctx, filter)
}

func (s *UserService) Update(ctx context.Context, id int64, input *domain.UpdateUserInput) (domain.User, error) {
	if id <= 0 {
		return domain.User{}, domain.ErrInvalidUserID
	}
	if err := input.Validate(); err != nil {
		return domain.User{}, err
	}

	if input.FirstName == nil && input.LastName == nil {
		return s.GetByID(ctx, id)
	}

	return s.repo.Update(ctx, id, input)
}

func (s *UserService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return domain.ErrInvalidUserID
	}
	return s.repo.Delete(ctx, id)
}
