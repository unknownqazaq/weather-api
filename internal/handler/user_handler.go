package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"weather-api/internal/auth"
	"weather-api/internal/dto"
	"weather-api/internal/model"
)

type UserService interface {
	Create(ctx context.Context, input *dto.CreateUserRequest) (model.User, error)
	GetByID(ctx context.Context, id int64) (model.User, error)
	List(ctx context.Context, filter dto.ListUsersFilter) ([]model.User, error)
	Update(ctx context.Context, id int64, input *dto.UpdateUserRequest) (model.User, error)
	Delete(ctx context.Context, id int64) error
}

type usersResponse struct {
	Data []dto.UserResponse `json:"data"`
}

type userResponse struct {
	Data dto.UserResponse `json:"data"`
}

type UserHandler struct {
	service UserService
}

func NewUserHandler(service UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.UserClaimsFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	user, err := h.service.GetByID(r.Context(), claims.UserID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.MapUser(user))
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input dto.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid json body"})
		return
	}
	user, err := h.service.Create(r.Context(), &input)
	if err != nil {
		h.handleError(w, err)
		return
	}
	w.Header().Set("Location", "/api/v1/users/"+strconv.FormatInt(user.ID, 10))
	writeJSON(w, http.StatusCreated, userResponse{Data: dto.MapUser(user)})

}
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	user, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, userResponse{Data: dto.MapUser(user)})
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := dto.ListUsersFilter{
		Limit:  parseIntQuery(r, "limit", 20),
		Offset: parseIntQuery(r, "offset", 0),
		Query:  r.URL.Query().Get("q"),
	}

	users, err := h.service.List(r.Context(), filter)
	if err != nil {
		h.handleError(w, err)
		return
	}

	var mappedUsers []dto.UserResponse
	for _, u := range users {
		mappedUsers = append(mappedUsers, dto.MapUser(u))
	}

	writeJSON(w, http.StatusOK, usersResponse{Data: mappedUsers})
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	var input dto.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid json body"})
		return
	}

	user, err := h.service.Update(r.Context(), id, &input)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, userResponse{Data: dto.MapUser(user)})
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
func (h *UserHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, model.ErrInvalidUserID), errors.Is(err, dto.ErrInvalidUserInput):
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	case errors.Is(err, model.ErrUserNotFound):
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
	case errors.Is(err, model.ErrEmailAlreadyTaken):
		writeJSON(w, http.StatusConflict, ErrorResponse{Error: err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
}
