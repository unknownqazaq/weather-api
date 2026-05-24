package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"weather-api/internal/dto"
	"weather-api/internal/handler"
	"weather-api/internal/model"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Create(ctx context.Context, input *dto.CreateUserRequest) (model.User, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(model.User), args.Error(1)
}

func (m *MockUserService) GetByID(ctx context.Context, id int64) (model.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(model.User), args.Error(1)
}

func (m *MockUserService) List(ctx context.Context, filter dto.ListUsersFilter) ([]model.User, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]model.User), args.Error(1)
}

func (m *MockUserService) Update(ctx context.Context, id int64, input *dto.UpdateUserRequest) (model.User, error) {
	args := m.Called(ctx, id, input)
	return args.Get(0).(model.User), args.Error(1)
}

func (m *MockUserService) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupRouter(userService *MockUserService) *chi.Mux {
	r := chi.NewRouter()
	userHandler := handler.NewUserHandler(userService)
	
	r.Post("/users", userHandler.Create)
	r.Get("/users/{id}", userHandler.GetByID)
	
	return r
}

func TestUserHandler_GetByID_Success(t *testing.T) {
	userService := new(MockUserService)
	router := setupRouter(userService)

	expectedUser := model.User{
		ID:        1,
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}

	userService.On("GetByID", mock.Anything, int64(1)).
		Return(expectedUser, nil).
		Once()

	req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)

	var result struct {
		Data dto.UserResponse `json:"data"`
	}
	err := json.Unmarshal(resp.Body.Bytes(), &result)

	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Data.ID)
	assert.Equal(t, "test@example.com", result.Data.Email)

	userService.AssertExpectations(t)
}

func TestUserHandler_GetByID_NotFound(t *testing.T) {
	userService := new(MockUserService)
	router := setupRouter(userService)

	userService.On("GetByID", mock.Anything, int64(999)).
		Return(model.User{}, model.ErrUserNotFound).
		Once()

	req := httptest.NewRequest(http.MethodGet, "/users/999", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusNotFound, resp.Code)
	userService.AssertExpectations(t)
}

func TestUserHandler_GetByID_BadRequest(t *testing.T) {
	userService := new(MockUserService)
	router := setupRouter(userService)

	req := httptest.NewRequest(http.MethodGet, "/users/abc", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	userService.AssertNotCalled(t, "GetByID")
}

func TestUserHandler_Create_Success(t *testing.T) {
	userService := new(MockUserService)
	router := setupRouter(userService)

	input := dto.CreateUserRequest{
		Email:        "test@example.com",
		PasswordHash: "password123",
		FirstName:    "Test",
		LastName:     "User",
	}

	expectedUser := model.User{
		ID:        2,
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}

	userService.On("Create", mock.Anything, &input).
		Return(expectedUser, nil).
		Once()

	body, err := json.Marshal(input)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusCreated, resp.Code)

	var result struct {
		Data dto.UserResponse `json:"data"`
	}
	err = json.Unmarshal(resp.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, int64(2), result.Data.ID)
	assert.Equal(t, "test@example.com", result.Data.Email)

	userService.AssertExpectations(t)
}

func TestUserHandler_Create_InvalidJSON(t *testing.T) {
	userService := new(MockUserService)
	router := setupRouter(userService)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString("{invalid_json"))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	userService.AssertNotCalled(t, "Create")
}
