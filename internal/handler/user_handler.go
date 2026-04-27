package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"weather-api/internal/domain"
)

type UserService interface {
	Create(ctx context.Context, input *domain.CreateUserInput) (domain.User, error)
	GetByID(ctx context.Context, id int64) (domain.User, error)
	List(ctx context.Context, filter domain.ListUsersFilter) ([]domain.User, error)
	Update(ctx context.Context, id int64, input *domain.UpdateUserInput) (domain.User, error)
	Delete(ctx context.Context, id int64) error
}

type usersResponse struct {
	Data []domain.User `json:"data"`
}

type userResponse struct {
	Data domain.User `json:"data"`
}

type UserHandler struct {
	service UserService
}

func NewUserHandler(service UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input domain.CreateUserInput
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
	writeJSON(w, http.StatusCreated, userResponse{Data: user})

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

	writeJSON(w, http.StatusOK, userResponse{Data: user})
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := domain.ListUsersFilter{
		Limit:  parseIntQuery(r, "limit", 20),
		Offset: parseIntQuery(r, "offset", 0),
		Query:  r.URL.Query().Get("q"),
	}

	users, err := h.service.List(r.Context(), filter)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, usersResponse{Data: users})
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	var input domain.UpdateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid json body"})
		return
	}

	user, err := h.service.Update(r.Context(), id, &input)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, userResponse{Data: user})
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
	case errors.Is(err, domain.ErrInvalidUserID), errors.Is(err, domain.ErrInvalidUserInput):
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrUserNotFound):
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrEmailAlreadyTaken):
		writeJSON(w, http.StatusConflict, ErrorResponse{Error: err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
}
