package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/mdflamingo/GophKeeper/internal/logger"
	"github.com/mdflamingo/GophKeeper/internal/model"
	"github.com/mdflamingo/GophKeeper/internal/server/service"
	"go.uber.org/zap"
)

// RegisterHandler godoc
// @Summary      Регистрация нового пользователя
// @Description  Создает нового пользователя с указанным логином и паролем
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body model.AuthUser true "Данные для регистрации"
// @Success      200 {object} model.AuthResponse "Успешная регистрация, токен в cookie и теле ответа"
// @Failure      400 {object} map[string]string "Неверный запрос или пустые поля"
// @Failure      409 {object} map[string]string "Пользователь с таким логином уже существует"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /user/register [post]
func RegisterHandler(w http.ResponseWriter, r *http.Request, svc *service.UserService, secretKey string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Log.Error("failed to read request body", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var req model.AuthUser
	if err := json.Unmarshal(body, &req); err != nil {
		logger.Log.Warn("invalid request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, err := svc.Register(req)
	if err != nil {
		handleRegistrationError(w, err, req.Login)
		return
	}

	token, err := svc.GenerateToken(userID, secretKey)
	if err != nil {
		logger.Log.Error("failed to create JWT token", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	setTokenCookie(w, token)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := model.AuthResponse{Token: token}
	respJSON, _ := json.Marshal(resp)
	w.Write(respJSON)
}

// handleRegistrationError обрабатывает ошибки регистрации
func handleRegistrationError(w http.ResponseWriter, err error, login string) {
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

// LoginHandler godoc
// @Summary      Аутентификация пользователя
// @Description  Вход в систему с получением JWT токена
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body model.AuthUser true "Учетные данные пользователя"
// @Success      200 {object} model.AuthResponse "Успешный вход, токен в cookie и теле ответа"
// @Failure      400 {object} map[string]string "Неверный запрос или пустые поля"
// @Failure      401 {object} map[string]string "Неверный логин или пароль"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /user/login [post]
func LoginHandler(w http.ResponseWriter, r *http.Request, svc *service.UserService, secretKey string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Log.Error("failed to read request body", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var req model.AuthUser
	if err := json.Unmarshal(body, &req); err != nil {
		logger.Log.Warn("invalid request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, err := svc.Login(req)
	if err != nil {
		handleLoginError(w, err, req.Login)
		return
	}

	token, err := svc.GenerateToken(userID, secretKey)
	if err != nil {
		logger.Log.Error("failed to create JWT token", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	setTokenCookie(w, token)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := model.AuthResponse{Token: token}
	respJSON, _ := json.Marshal(resp)
	w.Write(respJSON)
}

// handleLoginError обрабатывает ошибки входа
func handleLoginError(w http.ResponseWriter, err error, login string) {
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
func setTokenCookie(w http.ResponseWriter, token string) {
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
