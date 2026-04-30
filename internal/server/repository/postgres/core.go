package postgres

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mdflamingo/GophKeeper/internal/config"
	"github.com/mdflamingo/GophKeeper/internal/logger"
	"go.uber.org/zap"
)

type Storage interface {
	Save(shortURL, originalURL, userID string) (string, error)
	Get(shortURL string) (originalURL string, found bool, deleted bool)
	GetList(userID int) ([]UserDataDB, error)
	Delete(doneCh chan struct{}, inputCh chan string, userID string) chan error
	Close() error
	Ping(ctx context.Context) error
	SaveUser(user UserDB) (int, error)
	GetUser(user UserDB) (int, error)
}

type DBStorage struct {
	dsn      string
	pool     *pgxpool.Pool
	poolOnce sync.Once
	initErr  error
}

func NewDBStorage(dsn string) (*DBStorage, error) {
	storage := &DBStorage{
		dsn: dsn,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := storage.initPool(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize pool: %w", err)
	}

	return storage, nil
}

func (d *DBStorage) initPool(ctx context.Context) error {
	config, err := pgxpool.ParseConfig(d.dsn)
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	ctxPing, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctxPing); err != nil {
		pool.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	d.pool = pool

	if err := d.runMigrationsSync(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
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

func (d *DBStorage) Close() error {
	if d.pool != nil {
		d.pool.Close()
	}
	return nil
}

func (d *DBStorage) Ping(ctx context.Context) error {
	if d.pool == nil {
		return fmt.Errorf("database pool is not initialized")
	}
	return d.pool.Ping(ctx)
}
