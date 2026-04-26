package service

import (
	"context"
	"weather-api/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

type userRepository interface {
	Create(ctx context.Context, input *domain.CreateUserInput) (domain.User, error)
	GetByID(ctx context.Context, id int64) (domain.User, error)
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
