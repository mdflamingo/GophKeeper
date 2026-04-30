package handler

import (
	"errors"
	"net/http"
	"time"

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

// setTokenCookie устанавливает JWT токен в cookie
func SetTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(30 * 24 * time.Hour.Seconds()),
		Path:     "/",
	})
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
