package service

import (
	"context"
	"weather-api/internal/dto"
	"weather-api/internal/model"

	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	Create(ctx context.Context, input *dto.CreateUserRequest) (model.User, error)
	GetByID(ctx context.Context, id int64) (model.User, error)
	GetByEmail(ctx context.Context, email string) (model.User, error)
	List(ctx context.Context, filter dto.ListUsersFilter) ([]model.User, error)
	Update(ctx context.Context, id int64, input *dto.UpdateUserRequest) (model.User, error)
	Delete(ctx context.Context, id int64) error
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Create(ctx context.Context, input *dto.CreateUserRequest) (model.User, error) {
	if err := input.NormalizeAndValidate(); err != nil {
		return model.User{}, err
	}
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(input.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return model.User{}, err
	}
	input.PasswordHash = string(hashBytes)

	return s.repo.Create(ctx, input)
}

func (s *UserService) GetByID(ctx context.Context, id int64) (model.User, error) {
	if id <= 0 {
		return model.User{}, model.ErrInvalidUserID
	}
	return s.repo.GetByID(ctx, id)
}

func (s *UserService) List(ctx context.Context, filter dto.ListUsersFilter) ([]model.User, error) {
	filter.Normalize()
	return s.repo.List(ctx, filter)
}

func (s *UserService) Update(ctx context.Context, id int64, input *dto.UpdateUserRequest) (model.User, error) {
	if id <= 0 {
		return model.User{}, model.ErrInvalidUserID
	}
	if err := input.Validate(); err != nil {
		return model.User{}, err
	}

	if input.FirstName == nil && input.LastName == nil {
		return s.GetByID(ctx, id)
	}

	return s.repo.Update(ctx, id, input)
}

func (s *UserService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return model.ErrInvalidUserID
	}
	return s.repo.Delete(ctx, id)
}
