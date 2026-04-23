package postgres

import (
	"context"
	"net/http"
	"time"

	"github.com/mdflamingo/GophKeeper/internal/logger"
	"go.uber.org/zap"
)

// Close implements [Storage].
func (d *DBStorage) Close() error {
	panic("unimplemented")
}

// Delete implements [Storage].
func (d *DBStorage) Delete(doneCh chan struct{}, inputCh chan string, userID string) chan error {
	panic("unimplemented")
}

// Get implements [Storage].
func (d *DBStorage) Get(shortURL string) (originalURL string, found bool, deleted bool) {
	panic("unimplemented")
}

// GetList implements [Storage].
func (d *DBStorage) GetList() {
	panic("unimplemented")
}

// Ping implements [Storage].
func (d *DBStorage) Ping(ctx context.Context) error {
	panic("unimplemented")
}

// Save implements [Storage].
func (d *DBStorage) Save(shortURL string, originalURL string, userID string) (string, error) {
	panic("unimplemented")
}

func HealthCheck(response http.ResponseWriter, request *http.Request, storage Storage) {
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
