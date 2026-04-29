package service

import (
	"context"
	"net/http"
	"time"

	"github.com/mdflamingo/GophKeeper/internal/logger"
	"github.com/mdflamingo/GophKeeper/internal/server/repository/postgres"
	"go.uber.org/zap"
)

func HealthCheck(response http.ResponseWriter, request *http.Request, storage postgres.Storage) {
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
