package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"weather-api/internal/domain"

	"github.com/go-chi/chi/v5"
)

type UserService interface {
	Create(ctx context.Context, input *domain.CreateUserInput) (domain.User, error)
	GetByID(ctx context.Context, id int64) (domain.User, error)
}

type errorResponse struct {
	Error string `json:"error"`
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
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json body"})
		return
	}
	user, err := h.service.Create(r.Context(), &input)
	if err != nil {
		h.handleError(w, err)
		return
	}
	w.Header().Set("Location", "/api/v1/users/"+strconv.FormatInt(user.ID, 10))
	writeJSON(w, http.StatusCreated, user)

}
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid user id"})
		return
	}
	user, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, userResponse{Data: user})
}
func (h *UserHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidUserID), errors.Is(err, domain.ErrInvalidUserInput):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrUserNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrEmailAlreadyTaken):
		writeJSON(w, http.StatusConflict, errorResponse{Error: err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}
