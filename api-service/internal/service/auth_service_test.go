package service_test

import (
	"context"
	"testing"
	"time"
	"weather-api/internal/auth"
	"weather-api/internal/dto"
	"weather-api/internal/model"
	"weather-api/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_Register_Success(t *testing.T) {
	repo := new(MockUserRepository)
	userService := service.NewUserService(repo)
	jwtManager := auth.NewJWTManager("secret-key", 10*time.Minute)
	authService := service.NewAuthService(userService, repo, jwtManager)

	input := &dto.CreateUserRequest{
		Email:        "register@example.com",
		PasswordHash: "pass123",
		FirstName:    "First",
		LastName:     "Last",
		Role:         "user",
	}

	repo.On("Create", mock.Anything, mock.Anything).Return(model.User{
		ID:        1,
		Email:     "register@example.com",
		FirstName: "First",
		LastName:  "Last",
		Role:      "user",
	}, nil).Once()

	resp, err := authService.Register(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, "register@example.com", resp.Email)
	assert.Equal(t, int64(1), resp.ID)
	repo.AssertExpectations(t)
}

func TestAuthService_Login_Success(t *testing.T) {
	repo := new(MockUserRepository)
	userService := service.NewUserService(repo)
	jwtManager := auth.NewJWTManager("secret-key", 10*time.Minute)
	authService := service.NewAuthService(userService, repo, jwtManager)

	password := "pass123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	user := model.User{
		ID:           1,
		Email:        "login@example.com",
		PasswordHash: string(hash),
		Role:         "user",
	}

	repo.On("GetByEmail", mock.Anything, "login@example.com").Return(user, nil).Once()

	input := dto.LoginRequest{
		Email:    "login@example.com",
		Password: password,
	}

	resp, err := authService.Login(context.Background(), input)

	require.NoError(t, err)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, "login@example.com", resp.User.Email)
	repo.AssertExpectations(t)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	repo := new(MockUserRepository)
	userService := service.NewUserService(repo)
	jwtManager := auth.NewJWTManager("secret-key", 10*time.Minute)
	authService := service.NewAuthService(userService, repo, jwtManager)

	repo.On("GetByEmail", mock.Anything, "nonexistent@example.com").Return(model.User{}, model.ErrUserNotFound).Once()

	input := dto.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "pass",
	}

	resp, err := authService.Login(context.Background(), input)

	require.ErrorIs(t, err, model.ErrUserNotFound)
	assert.Nil(t, resp)
	repo.AssertExpectations(t)
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	repo := new(MockUserRepository)
	userService := service.NewUserService(repo)
	jwtManager := auth.NewJWTManager("secret-key", 10*time.Minute)
	authService := service.NewAuthService(userService, repo, jwtManager)

	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	require.NoError(t, err)

	user := model.User{
		ID:           1,
		Email:        "login@example.com",
		PasswordHash: string(hash),
		Role:         "user",
	}

	repo.On("GetByEmail", mock.Anything, "login@example.com").Return(user, nil).Once()

	input := dto.LoginRequest{
		Email:    "login@example.com",
		Password: "wrong-password",
	}

	resp, err := authService.Login(context.Background(), input)

	require.ErrorIs(t, err, model.ErrUserNotFound)
	assert.Nil(t, resp)
	repo.AssertExpectations(t)
}
