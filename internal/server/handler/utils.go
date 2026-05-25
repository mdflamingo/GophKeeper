package handler

import (
	"errors"
	"net/http"

	"github.com/mdflamingo/GophKeeper/internal/logger"
	"github.com/mdflamingo/GophKeeper/internal/server/service"
	"go.uber.org/zap"
)

// handleLoginError обрабатывает ошибки входа
func HandleLoginError(w http.ResponseWriter, err error, login string) {
	switch {
	case errors.Is(err, service.ErrEmptyRequiredField):
		logger.Log.Warn("empty required field", zap.String("login", login))
		http.Error(w, "Login and password are required", http.StatusBadRequest)
	case errors.Is(err, service.ErrInvalidCredentials):
		logger.Log.Warn("invalid credentials", zap.String("login", login))
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
	default:
		logger.Log.Error("failed to login user", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// handleRegistrationError обрабатывает ошибки регистрации
func HandleRegistrationError(w http.ResponseWriter, err error, login string) {
	switch {
	case errors.Is(err, service.ErrEmptyRequiredField):
		logger.Log.Warn("empty required field", zap.String("login", login))
		http.Error(w, "Login and password are required", http.StatusBadRequest)
	case errors.Is(err, service.ErrUserAlreadyExists):
		logger.Log.Warn("user already exists", zap.String("login", login))
		http.Error(w, "User already exists", http.StatusConflict)
	default:
		logger.Log.Error("failed to register user", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
