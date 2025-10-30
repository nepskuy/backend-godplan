package utils

import "net/http"

type AppError struct {
	Code    int
	Message string
}

func (e AppError) Error() string {
	return e.Message
}

var (
	ErrInvalidCredentials = AppError{Code: http.StatusUnauthorized, Message: "Invalid credentials"}
	ErrUserNotFound       = AppError{Code: http.StatusNotFound, Message: "User not found"}
	ErrTaskNotFound       = AppError{Code: http.StatusNotFound, Message: "Task not found"}
	ErrInternalServer     = AppError{Code: http.StatusInternalServerError, Message: "Internal server error"}
	ErrEmailExists        = AppError{Code: http.StatusBadRequest, Message: "Email already exists"}
)
