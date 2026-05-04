package service

import (
	"context"
	"weather-api/internal/auth"
	"weather-api/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userService *UserService
	userRepo    UserRepository
	jwtManager  *auth.JWTManager
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  domain.User `json:"user"`
}

func NewAuthService(userService *UserService, userRepo UserRepository, jwtManager *auth.JWTManager) *AuthService {
	return &AuthService{
		userService: userService,
		userRepo:    userRepo,
		jwtManager:  jwtManager,
	}
}

func (s *AuthService) Register(ctx context.Context, input *domain.CreateUserInput) (domain.User, error) {
	// The user service hashes the password and creates the user
	return s.userService.Create(ctx, input)
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (*AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password))
	if err != nil {
		return nil, domain.ErrUserNotFound // To not expose whether the email or password was wrong
	}

	token, err := s.jwtManager.Generate(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  user,
	}, nil
}
