package postgres

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mdflamingo/GophKeeper/internal/config"
	"github.com/mdflamingo/GophKeeper/internal/logger"
	"go.uber.org/zap"
)

type Storage interface {
	Save(shortURL, originalURL, userID string) (string, error)
	Get(shortURL string) (originalURL string, found bool, deleted bool)
	GetList()
	Delete(doneCh chan struct{}, inputCh chan string, userID string) chan error
	Close() error
	Ping(ctx context.Context) error
}

type DBStorage struct {
	dsn      string
	pool     *pgxpool.Pool
	poolOnce sync.Once
	initErr  error
}

func NewDBStorage(dsn string) (*DBStorage, error) {
	return &DBStorage{
		dsn: dsn,
	}, nil
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
	if dataBaseDSN != "" {
		if storage, err := NewDBStorage(dataBaseDSN); err == nil {
			logger.Log.Info("Successfully initialized database storage")
			return storage, nil
		} else {
			logger.Log.Error("Failed to initialize database storage", zap.Error(err))
			return nil, err
		}
	}

	return nil, errors.New("database DSN is empty")
}
