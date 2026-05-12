package model

import "errors"

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidUserID     = errors.New("invalid user id")
	ErrEmailAlreadyTaken = errors.New("email already exists")
)
