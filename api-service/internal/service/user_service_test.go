package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"weather-api/internal/dto"
	"weather-api/internal/model"
	"weather-api/internal/service"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) (model.User, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(model.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int64) (model.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(model.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (model.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(model.User), args.Error(1)
}

func (m *MockUserRepository) List(ctx context.Context, filter dto.ListUsersFilter) ([]model.User, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]model.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, id int64, input *dto.UpdateUserRequest) (model.User, error) {
	args := m.Called(ctx, id, input)
	return args.Get(0).(model.User), args.Error(1)
}

func (m *MockUserRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestUserService_GetByID_Success(t *testing.T) {
	repo := new(MockUserRepository)
	userService := service.NewUserService(repo)

	expectedUser := model.User{
		ID:        1,
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}

	repo.On("GetByID", mock.Anything, int64(1)).
		Return(expectedUser, nil).
		Once()

	actualUser, err := userService.GetByID(context.Background(), 1)

	require.NoError(t, err)
	assert.Equal(t, expectedUser, actualUser)

	repo.AssertExpectations(t)
}

func TestUserService_GetByID_InvalidID(t *testing.T) {
	repo := new(MockUserRepository)
	userService := service.NewUserService(repo)

	user, err := userService.GetByID(context.Background(), 0)

	require.ErrorIs(t, err, model.ErrInvalidUserID)
	assert.Empty(t, user)

	repo.AssertNotCalled(t, "GetByID")
}

func TestUserService_GetByID_NotFound(t *testing.T) {
	repo := new(MockUserRepository)
	userService := service.NewUserService(repo)

	repo.On("GetByID", mock.Anything, int64(999)).
		Return(model.User{}, model.ErrUserNotFound).
		Once()

	user, err := userService.GetByID(context.Background(), 999)

	require.ErrorIs(t, err, model.ErrUserNotFound)
	assert.Empty(t, user)

	repo.AssertExpectations(t)
}

func TestUserService_Delete_Success(t *testing.T) {
	repo := new(MockUserRepository)
	userService := service.NewUserService(repo)

	repo.On("Delete", mock.Anything, int64(1)).
		Return(nil).
		Once()

	err := userService.Delete(context.Background(), 1)

	require.NoError(t, err)

	repo.AssertExpectations(t)
}

func TestUserService_Delete_InvalidID(t *testing.T) {
	repo := new(MockUserRepository)
	userService := service.NewUserService(repo)

	err := userService.Delete(context.Background(), -1)

	require.ErrorIs(t, err, model.ErrInvalidUserID)

	repo.AssertNotCalled(t, "Delete")
}

func TestUserService_Create_Success(t *testing.T) {
	repo := new(MockUserRepository)
	userService := service.NewUserService(repo)

	input := &dto.CreateUserRequest{
		Email:        "test@example.com",
		PasswordHash: "password123",
		FirstName:    "Test",
		LastName:     "User",
		Role:         "user",
	}

	repo.On("Create", mock.Anything, mock.MatchedBy(func(u *model.User) bool {
		return u.Email == "test@example.com" && u.FirstName == "Test"
	})).Return(model.User{ID: 1, Email: "test@example.com"}, nil).Once()

	user, err := userService.Create(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, int64(1), user.ID)

	repo.AssertExpectations(t)
}

func TestUserService_Create_InvalidInput(t *testing.T) {
	repo := new(MockUserRepository)
	userService := service.NewUserService(repo)

	input := &dto.CreateUserRequest{
		Email: "invalid-email",
	}

	_, err := userService.Create(context.Background(), input)

	require.Error(t, err)
	repo.AssertNotCalled(t, "Create")
}

func TestUserService_List_Success(t *testing.T) {
	repo := new(MockUserRepository)
	userService := service.NewUserService(repo)

	filter := dto.ListUsersFilter{Limit: 10}
	
	repo.On("List", mock.Anything, mock.Anything).Return([]model.User{
		{ID: 1, Email: "test1@example.com"},
	}, nil).Once()

	users, err := userService.List(context.Background(), filter)

	require.NoError(t, err)
	assert.Len(t, users, 1)

	repo.AssertExpectations(t)
}

func TestUserService_Update_Success(t *testing.T) {
	repo := new(MockUserRepository)
	userService := service.NewUserService(repo)

	firstName := "Updated"
	input := &dto.UpdateUserRequest{
		FirstName: &firstName,
	}

	repo.On("Update", mock.Anything, int64(1), input).Return(model.User{ID: 1, FirstName: "Updated"}, nil).Once()

	user, err := userService.Update(context.Background(), 1, input)

	require.NoError(t, err)
	assert.Equal(t, "Updated", user.FirstName)

	repo.AssertExpectations(t)
}

func TestUserService_Update_InvalidID(t *testing.T) {
	repo := new(MockUserRepository)
	userService := service.NewUserService(repo)

	_, err := userService.Update(context.Background(), 0, nil)

	require.ErrorIs(t, err, model.ErrInvalidUserID)
}

