package handler

import (
	"context"
	"encoding/json"
	"errors"
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
// @Failure      400 "Неверный запрос или пустые поля"
// @Failure      409 "Пользователь с таким логином уже существует"
// @Failure      500 "Внутренняя ошибка сервера"
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
// @Failure      400 "Неверный запрос или пустые поля"
// @Failure      401 "Неверный логин или пароль"
// @Failure      500 "Внутренняя ошибка сервера"
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
// @Security     BearerAuth
// @Success      200 {object} model.SecretListResponse "Успешное получение списка секретов"
// @Success      204 "Нет содержимого (список секретов пуст)"
// @Failure      401 "Пользователь не авторизован"
// @Failure      500 "Внутренняя ошибка сервера"
// @Router       /secret/list [get]
func GetSecretListHandler(w http.ResponseWriter, r *http.Request, svc *service.GopheKeeperService, bucketName string) {
	userID, err := GetUserIDFromRequest(r)
	if err != nil {
		logger.Log.Warn("failed to get user ID", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	secrets, err := svc.GetSecrets(userID, bucketName)
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
// @Security     BearerAuth
// @Param        id path int true "ID секрета"
// @Success      200 {object} model.SecretResponse "Успешное получение секрета"
// @Failure      204 "Секрет не найден"
// @Failure      400 "Неверный запрос"
// @Failure      401 "Пользователь не авторизован"
// @Failure      500 "Внутренняя ошибка сервера"
// @Router       /secret/{id} [get]
func GetSecretHandler(w http.ResponseWriter, r *http.Request, svc *service.GopheKeeperService, bucketName string) {
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

	item, err := svc.GetOneSecret(userID, secretID, bucketName)
	if err != nil {
		if errors.Is(err, service.ErrSecretNotFound) {
			logger.Log.Error("secret not found", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusNoContent), http.StatusNoContent)
			return
		} else {
			logger.Log.Error("failed to get user secrets", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

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

// SaveSecretHandler godoc
// @Summary      Создание секрета пользователя
// @Description  Создает секрет текущего авторизованного пользователя
// @Param        request body model.SecretCreateRequest true "Данные секрета"
// @Tags         secrets
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      201  "Успешное создание секрета"
// @Failure      400  "Неверный запрос"
// @Failure      401  "Пользователь не авторизован"
// @Failure      500  "Внутренняя ошибка сервера"
// @Router       /secret [post]
func SaveSecretHandler(w http.ResponseWriter, r *http.Request, svc *service.GopheKeeperService) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Log.Error("failed to read request body", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var inputSecret model.SecretCreateRequest
	if err := json.Unmarshal(body, &inputSecret); err != nil {
		logger.Log.Warn("invalid request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, err := GetUserIDFromRequest(r)
	if err != nil {
		logger.Log.Warn("failed to get user ID", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	err = svc.SaveOneSecret(inputSecret, userID)

	if err != nil {
		logger.Log.Error("failed to save secret", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

// SaveFileSecretHandler godoc
// @Summary      Загрузка файлового секрета
// @Description  Загружает файл как секрет пользователя
// @Tags         secrets
// @Accept       multipart/form-data
// @Security     BearerAuth
// @Param        file formData file true "Файл для загрузки"
// @Param        data_type formData string true "Тип секрета" Enums(TEXT, CARD, FILE, CREDENTIALS) default(FILE)
// @Param        data formData string true "Данные секрета в JSON формате" example({"login":"user","password":"pass"})
// @Success      201  "Успешная загрузка файла"
// @Failure      400  "Неверный запрос"
// @Failure      401  "Пользователь не авторизован"
// @Failure      413  "Файл слишком большой"
// @Failure      500  "Внутренняя ошибка сервера"
// @Router       /secret/file [post]
func SaveFileSecretHandler(w http.ResponseWriter, r *http.Request, gophekeeperService *service.GopheKeeperService, bucketName string) {
	const maxFileSize = 32 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxFileSize)

	err := r.ParseMultipartForm(maxFileSize)
	if err != nil {
		logger.Log.Error("failed to parse multipart form", zap.Error(err))
		http.Error(w, "file too large or invalid form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		logger.Log.Error("failed to get file from form", zap.Error(err))
		http.Error(w, "failed to get file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	userID, err := GetUserIDFromRequest(r)
	if err != nil {
		logger.Log.Warn("failed to get user ID", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	dataType := r.FormValue("data_type")
	dataJSON := r.FormValue("data")

	inputSecret := model.SecretCreateRequest{
		DataType: model.DataType(dataType),
		Data:     json.RawMessage(dataJSON),
	}

	err = gophekeeperService.SaveFile(
		r.Context(),
		file,
		header.Filename,
		userID,
		bucketName,
		header.Size,
		inputSecret,
	)
	if err != nil {
		logger.Log.Error("failed to save file", zap.Error(err))
		http.Error(w, "failed to save file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
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
