package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mdflamingo/GophKeeper/internal/logger"
	"github.com/mdflamingo/GophKeeper/internal/model"
)

var ErrConflict = errors.New("conflict: duplicate entry")
var ErrNotFound = errors.New("obj not found")

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

func (d *DBStorage) runMigrationsSync() error {
	logger.Log.Info("Running database migrations (sync)")

	db, err := sql.Open("postgres", d.dsn)
	if err != nil {
		return fmt.Errorf("failed to open database for migrations: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database for migrations: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Log.Info("Migrations completed successfully")
	return nil
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

func (d *DBStorage) getPool(ctx context.Context) (*pgxpool.Pool, error) {
	d.poolOnce.Do(func() {
		config, err := pgxpool.ParseConfig(d.dsn)
		if err != nil {
			d.initErr = fmt.Errorf("failed to parse config: %w", err)
			return
		}

		config.MaxConns = 10
		config.MinConns = 2
		config.MaxConnLifetime = time.Hour
		config.MaxConnIdleTime = 30 * time.Minute
		config.HealthCheckPeriod = time.Minute

		config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeDescribeExec

		pool, err := pgxpool.NewWithConfig(ctx, config)
		if err != nil {
			d.initErr = fmt.Errorf("failed to create connection pool: %w", err)
			return
		}

		ctxPing, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if err := pool.Ping(ctxPing); err != nil {
			pool.Close()
			d.initErr = fmt.Errorf("failed to ping database: %w", err)
			return
		}

		d.pool = pool

		if err := d.runMigrationsSync(); err != nil {
			d.initErr = fmt.Errorf("failed to run migrations: %w", err)
			return
		}
	})

	if d.initErr != nil {
		return nil, d.initErr
	}
	return d.pool, nil
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

///
// Querys
///

func (d *DBStorage) Get(userID, secrertID int) (SecretDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var secret SecretDB

	err := d.pool.QueryRow(
		ctx,
		`SELECT id, data_type, metadata, created_at FROM user_data WHERE user_id = $1 and id = $2`,
		userID, secrertID).Scan(&secret.ID, &secret.DataType, &secret.MetaData, &secret.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SecretDB{}, ErrNotFound
		}
		return SecretDB{}, fmt.Errorf("failed to get secret: %w", err)
	}
	return secret, nil
}

func (d *DBStorage) GetList(userID int) ([]SecretDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := d.getPool(ctx)
	if err != nil {
		return nil, fmt.Errorf("database not available: %w", err)
	}

	rows, err := pool.Query(ctx,
		"SELECT id, data_type, metadata, created_at FROM user_data WHERE user_id = $1",
		userID)

	if err != nil {
		return nil, fmt.Errorf("database query error: %w", err)
	}
	defer rows.Close()

	var items []SecretDB

	for rows.Next() {
		var item SecretDB

		err := rows.Scan(&item.ID, &item.DataType, &item.MetaData, &item.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("data scan error: %w", err)
		}

		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows processing error: %w", err)
	}

	return items, nil
}

func (d *DBStorage) Save(secret model.SecretCreateRequest, userID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := d.pool.Exec(ctx,
		`INSERT INTO user_data (user_id, data_type, metadata)
         VALUES ($1, $2, $3)`,
		userID, secret.DataType, secret.Data)

	if err != nil {
		return fmt.Errorf("failed to save secret: %w", err)
	}

	return nil
}

func (d *DBStorage) SaveUser(user UserDB) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var userID int

	err := d.pool.QueryRow(ctx,
		`INSERT INTO users (login, password)
         VALUES ($1, $2)
		 RETURNING id`,
		user.Login, user.Password).Scan(&userID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return 0, ErrConflict
		}
		return 0, fmt.Errorf("failed to save user: %w", err)
	}

	return userID, nil
}

func (d *DBStorage) GetUser(user UserDB) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var userID int

	err := d.pool.QueryRow(ctx, `SELECT id FROM users WHERE login = $1 and password = $2`, user.Login, user.Password).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrNotFound
		}
		return 0, fmt.Errorf("failed to get user: %w", err)
	}
	return userID, nil
}
