package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mdflamingo/GophKeeper/internal/logger"
	"go.uber.org/zap"
)

type contextKey string

const userIDKey contextKey = "userID"

func AuthMiddleware(secretKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				logger.Log.Debug("no authorization header found")
				http.Error(w, "Unauthorized: missing authorization header", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			userID, err := validateJWT(tokenString, secretKey)
			if err != nil {
				logger.Log.Warn("invalid token", zap.Error(err))
				http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func validateJWT(tokenString, secretKey string) (int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid token claims")
	}

	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return 0, errors.New("token expired")
		}
	}

	userIDFloat, ok := claims["userID"].(float64)
	if !ok {
		return 0, errors.New("userID not found in token")
	}

	return int(userIDFloat), nil
}

func GetUserIDFromRequest(r *http.Request) (int, error) {
	ctx := r.Context()
	userIDValue := ctx.Value(userIDKey)
	if userIDValue == nil {
		return 0, errors.New("userID not found in context")
	}

	userID, ok := userIDValue.(int)
	if !ok {
		return 0, errors.New("userID is not an integer")
	}

	return userID, nil
}
