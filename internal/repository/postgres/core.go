package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/mdflamingo/GophKeeper/internal/config"
	"github.com/mdflamingo/GophKeeper/internal/logger"
	"github.com/mdflamingo/GophKeeper/internal/model"
	"go.uber.org/zap"
)

type Storage interface {
	Save(secret model.SecretCreateRequest, userID int) (int, error)
	Get(userID, secretID int) (SecretDB, error)
	GetList(userID int) ([]SecretDB, error)
	Update(secretID, userID int, secret model.SecretUpdateRequest) error
	Close() error
	Ping(ctx context.Context) error
	SaveUser(user UserDB) (int, error)
	GetUser(user UserDB) (int, error)
}

func ConnectPG(pgConf *config.Postgres) (Storage, error) {
	logger.Log.Info("Starting initialize postgres storage")
	dataBaseDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		pgConf.PostgresUser,
		pgConf.PostgresPassword,
		pgConf.PostgresHost,
		pgConf.PostgresPort,
		pgConf.PostgresDB,
	)

	if dataBaseDSN == "" {
		return nil, errors.New("database DSN is empty")
	}

	storage, err := NewDBStorage(dataBaseDSN)
	if err != nil {
		logger.Log.Error("Failed to initialize database storage", zap.Error(err))
		return nil, err
	}

	logger.Log.Info("Successfully initialized database storage")
	return storage, nil
}
