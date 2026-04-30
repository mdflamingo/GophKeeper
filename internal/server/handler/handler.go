package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mdflamingo/GophKeeper/internal/logger"
	"github.com/mdflamingo/GophKeeper/internal/model"
	"github.com/mdflamingo/GophKeeper/internal/server/repository/postgres"
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
		HandleRegistrationError(w, err, req.Login)
		return
	}

	token, err := svc.GenerateToken(userID, secretKey)
	if err != nil {
		logger.Log.Error("failed to create JWT token", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	SetTokenCookie(w, token)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := model.AuthResponse{Token: token}
	respJSON, _ := json.Marshal(resp)
	w.Write(respJSON)
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
		HandleLoginError(w, err, req.Login)
		return
	}

	token, err := svc.GenerateToken(userID, secretKey)
	if err != nil {
		logger.Log.Error("failed to create JWT token", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	SetTokenCookie(w, token)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := model.AuthResponse{Token: token}
	respJSON, _ := json.Marshal(resp)
	w.Write(respJSON)
}

// GetSecretListHandler godoc
// @Summary      Получение списка секретов пользователя
// @Description  Возвращает список всех сохраненных секретов (данных) текущего авторизованного пользователя
// @Tags         secrets
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200 {array} postgres.UserDataDB "Успешное получение списка секретов"
// @Success      204 "Нет содержимого (список секретов пуст)"
// @Failure      400 {object} map[string]string "Неверный запрос"
// @Failure      401 {object} map[string]string "Пользователь не авторизован"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /secret/list [get]
func GetSecretListHandler(w http.ResponseWriter, r *http.Request, svc *service.GopheKeeperService) {
	userID, err := GetUserIDFromRequest(r)
	if err != nil {
		logger.Log.Warn("failed to get user ID", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	secrets, err := svc.GetSecrets(userID)
	if err != nil {
		logger.Log.Error("failed to get user secrets", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if secrets.Count == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	respJSON, err := json.Marshal(secrets)
	if err != nil {
		logger.Log.Error("failed to marshal response to JSON", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(respJSON)
}

// GetSecretHandler godoc
// @Summary      Получение секрета пользователя
// @Description  Возвращает секрет текущего авторизованного пользователя по id секрета
// @Tags         secrets
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  postgres.UserDataDB "Успешное получение секрета"
// @Success      204 "Секрет не найден"
// @Failure      400 {object} map[string]string "Неверный запрос"
// @Failure      401 {object} map[string]string "Пользователь не авторизован"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /secret [get]
func GetSecretHandler(w http.ResponseWriter, r *http.Request, svc *service.GopheKeeperService) {
	userID, err := GetUserIDFromRequest(r)
	if err != nil {
		logger.Log.Warn("failed to get user ID", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	secretIDStr := chi.URLParam(r, "id")
	if secretIDStr == "" {
		logger.Log.Warn("order number is empty")
		http.Error(w, "Order number is required", http.StatusBadRequest)
		return
	}
	secretID, err := strconv.Atoi(secretIDStr)
	if err != nil || secretID <= 0 {
		logger.Log.Warn("invalid secret ID", zap.String("id", secretIDStr))
		http.Error(w, "Invalid secret ID", http.StatusBadRequest)
		return
	}

	item, err := svc.GetOneSecret(userID, secretID)
	if err != nil {
		logger.Log.Error("failed to get user secrets", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	respJSON, err := json.Marshal(item)
	if err != nil {
		logger.Log.Error("failed to marshal response to JSON", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(respJSON)
}

// HealthCheck godoc
// @Summary      Проверка здоровья сервера
// @Description  Проверяет доступность сервера и подключение к БД
// @Tags         monitoring
// @Produce      plain
// @Success      200 {string} string "OK"
// @Failure      500 {string} string "Internal Server Error"
// @Router       /ping [get]
func DBHealthCheck(response http.ResponseWriter, request *http.Request, storage postgres.Storage) {
	logger.Log.Info("HealthCheck called", zap.String("method", request.Method))

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := storage.Ping(ctx); err != nil {
		logger.Log.Error("storage not available", zap.Error(err))
		http.Error(response, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	logger.Log.Info("HealthCheck completed successfully")
	response.WriteHeader(http.StatusOK)
	response.Write([]byte("OK"))
}
