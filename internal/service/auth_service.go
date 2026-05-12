package service

import (
	"context"
	"weather-api/internal/auth"
	"weather-api/internal/dto"
	"weather-api/internal/model"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userService *UserService
	userRepo    UserRepository
	jwtManager  *auth.JWTManager
}

func NewAuthService(userService *UserService, userRepo UserRepository, jwtManager *auth.JWTManager) *AuthService {
	return &AuthService{
		userService: userService,
		userRepo:    userRepo,
		jwtManager:  jwtManager,
	}
}

func (s *AuthService) Register(ctx context.Context, input *dto.CreateUserRequest) (dto.UserResponse, error) {
	// The user service hashes the password and creates the user
	user, err := s.userService.Create(ctx, input)
	if err != nil {
		return dto.UserResponse{}, err
	}
	return dto.MapUser(user), nil
}

func (s *AuthService) Login(ctx context.Context, input dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, model.ErrUserNotFound
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password))
	if err != nil {
		return nil, model.ErrUserNotFound // To not expose whether the email or password was wrong
	}

	token, err := s.jwtManager.Generate(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: token,
		User:  dto.MapUser(user),
	}, nil
}
