package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	apiModel "github.com/mdflamingo/GophKeeper/internal/model"
)

type SyncStatus string

const (
	SyncStatusPending SyncStatus = "pending"
	SyncStatusSynced  SyncStatus = "synced"
	SyncStatusFailed  SyncStatus = "failed"
	SyncStatusDeleted SyncStatus = "deleted"
)

type LocalSecret struct {
	ID            int        `json:"id"`
	ServerID      *int       `json:"server_id"`
	DataType      string     `json:"data_type"`
	DataJSON      string     `json:"data_json"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	SyncStatus    SyncStatus `json:"sync_status"`
	SyncAttempts  int        `json:"sync_attempts"`
	LastSyncError string     `json:"last_sync_error,omitempty"`
	LocalVersion  int        `json:"local_version"`
	IsDeleted     bool       `json:"is_deleted"`
}

type LocalStorage struct {
	db *sql.DB
}

func New(baseDir string) (*LocalStorage, error) {
	dir := filepath.Join(baseDir, ".data")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("создание директории: %w", err)
	}

	dbPath := filepath.Join(dir, "secrets.db")

	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_synchronous=NORMAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("открытие БД: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	if err := initTables(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("создание таблиц: %w", err)
	}

	return &LocalStorage{db: db}, nil
}

func initTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS local_secrets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			server_id INTEGER,
			data_type TEXT NOT NULL,
			data_json TEXT NOT NULL,
			metadata TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			sync_status TEXT DEFAULT 'pending',
			sync_attempts INTEGER DEFAULT 0,
			last_sync_error TEXT DEFAULT '',
			local_version INTEGER DEFAULT 1,
			is_deleted INTEGER DEFAULT 0
		)`,

		`CREATE INDEX IF NOT EXISTS idx_local_id ON local_secrets(id)`,
		`CREATE INDEX IF NOT EXISTS idx_server_id ON local_secrets(server_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sync_status ON local_secrets(sync_status)`,
		`CREATE INDEX IF NOT EXISTS idx_updated_at ON local_secrets(updated_at)`,
		`CREATE INDEX IF NOT EXISTS idx_data_type ON local_secrets(data_type)`,

		`CREATE TABLE IF NOT EXISTS sync_metadata (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("ошибка выполнения запроса: %w\nЗапрос: %s", err, query)
		}
	}

	return nil
}

func (s *LocalStorage) Close() error {
	return s.db.Close()
}

func (s *LocalStorage) SaveSecret(dataType apiModel.DataType, data any) (int, error) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return 0, fmt.Errorf("json marshal: %w", err)
	}

	now := time.Now()
	result, err := s.db.Exec(`
		INSERT INTO local_secrets (data_type, data_json, sync_status, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?)`,
		string(dataType),
		string(dataJSON),
		SyncStatusPending,
		now,
		now,
	)
	if err != nil {
		return 0, fmt.Errorf("insert secret: %w", err)
	}

	id64, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}
	id := int(id64)
	return id, nil
}

func (s *LocalStorage) SaveSecretWithMetadata(dataType apiModel.DataType, data any, metadata string) (int64, error) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return 0, fmt.Errorf("json marshal: %w", err)
	}

	now := time.Now()
	result, err := s.db.Exec(`
		INSERT INTO local_secrets (data_type, data_json, metadata, sync_status, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?)`,
		string(dataType),
		string(dataJSON),
		metadata,
		SyncStatusPending,
		now,
		now,
	)
	if err != nil {
		return 0, fmt.Errorf("insert secret: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}

	return id, nil
}

func (s *LocalStorage) UpdateSecret(localID int, dataType apiModel.DataType, data interface{}) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}

	result, err := s.db.Exec(`
		UPDATE local_secrets 
		SET data_type = ?, 
		    data_json = ?, 
		    sync_status = ?, 
		    updated_at = CURRENT_TIMESTAMP,
		    local_version = local_version + 1,
		    sync_attempts = 0,
		    last_sync_error = ''
		WHERE id = ? AND is_deleted = 0`,
		string(dataType),
		string(dataJSON),
		SyncStatusPending,
		localID,
	)
	if err != nil {
		return fmt.Errorf("update secret: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("локальный секрет #%d не найден", localID)
	}

	return nil
}

func (s *LocalStorage) MarkAsSynced(localID, serverID int) error {
	_, err := s.db.Exec(`
		UPDATE local_secrets 
		SET sync_status = ?, 
		    server_id = ?, 
		    sync_attempts = 0, 
		    last_sync_error = '',
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		SyncStatusSynced,
		serverID,
		localID,
	)
	return err
}

func (s *LocalStorage) GetSecretByID(localID int) (*LocalSecret, error) {
	query := `
		SELECT id, server_id, data_type, data_json, metadata,
		       created_at, updated_at, sync_status, sync_attempts, 
		       last_sync_error, local_version, is_deleted
		FROM local_secrets 
		WHERE id = ?`

	var secret LocalSecret
	var serverID sql.NullInt64
	var lastError sql.NullString
	var metadata sql.NullString

	err := s.db.QueryRow(query, localID).Scan(
		&secret.ID, &serverID, &secret.DataType, &secret.DataJSON, &metadata,
		&secret.CreatedAt, &secret.UpdatedAt, &secret.SyncStatus, &secret.SyncAttempts,
		&lastError, &secret.LocalVersion, &secret.IsDeleted,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if serverID.Valid {
		sid := int(serverID.Int64)
		secret.ServerID = &sid
	}
	if lastError.Valid {
		secret.LastSyncError = lastError.String
	}

	return &secret, nil
}

func (s *LocalStorage) GetSecretByServerID(serverID int64) (*LocalSecret, error) {
	query := `
		SELECT id, server_id, data_type, data_json, metadata,
		       created_at, updated_at, sync_status, sync_attempts, 
		       last_sync_error, local_version, is_deleted
		FROM local_secrets 
		WHERE server_id = ? AND is_deleted = 0`

	var secret LocalSecret
	var sid sql.NullInt64
	var lastError sql.NullString
	var metadata sql.NullString

	err := s.db.QueryRow(query, serverID).Scan(
		&secret.ID, &sid, &secret.DataType, &secret.DataJSON, &metadata,
		&secret.CreatedAt, &secret.UpdatedAt, &secret.SyncStatus, &secret.SyncAttempts,
		&lastError, &secret.LocalVersion, &secret.IsDeleted,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if sid.Valid {
		sidInt := int(sid.Int64)
		secret.ServerID = &sidInt
	}
	if lastError.Valid {
		secret.LastSyncError = lastError.String
	}

	return &secret, nil
}

func (s *LocalStorage) ListLocalSecrets() error {
	rows, err := s.db.Query(`
		SELECT id, server_id, data_type, sync_status, created_at, 
		       updated_at, sync_attempts, last_sync_error, is_deleted
		FROM local_secrets 
		WHERE is_deleted = 0
		ORDER BY updated_at DESC`)
	if err != nil {
		return fmt.Errorf("list secrets: %w", err)
	}
	defer rows.Close()

	count := 0

	for rows.Next() {
		count++
		var (
			localID      int64
			serverID     sql.NullInt64
			dataType     string
			syncStatus   string
			createdAt    time.Time
			updatedAt    time.Time
			syncAttempts int
			lastError    sql.NullString
			isDeleted    bool
		)

		err := rows.Scan(&localID, &serverID, &dataType, &syncStatus,
			&createdAt, &updatedAt, &syncAttempts, &lastError, &isDeleted)
		if err != nil {
			fmt.Printf("⚠️  Ошибка чтения строки: %v\n", err)
			continue
		}

		status := getStatusEmoji(syncStatus)
		serverStr := "нет"
		if serverID.Valid {
			serverStr = fmt.Sprintf("#%d", serverID.Int64)
		}

		fmt.Printf("[%d] %s %-3s | Тип: %-12s | Сервер: %-8s | Обновлен: %s",
			localID, status, syncStatus, dataType, serverStr,
			updatedAt.Format("02.01 15:04"))

		if syncStatus == string(SyncStatusFailed) && lastError.Valid {
			fmt.Printf("\n     ⚠️  Ошибка: %s (попыток: %d)", lastError.String, syncAttempts)
		}
		fmt.Println()
	}

	if count == 0 {
		fmt.Println("📭 Нет локальных секретов")
	} else {
		fmt.Printf("\n📊 Всего секретов: %d\n", count)
	}

	return nil
}

func getStatusEmoji(status string) string {
	switch SyncStatus(status) {
	case SyncStatusSynced:
		return "✅"
	case SyncStatusPending:
		return "⏳"
	case SyncStatusFailed:
		return "❌"
	case SyncStatusDeleted:
		return "🗑️"
	default:
		return "❓"
	}
}

func (s *LocalStorage) GetPendingSecrets(limit int) ([]*LocalSecret, error) {
	query := `
		SELECT id, server_id, data_type, data_json, metadata,
		       created_at, updated_at, sync_status, sync_attempts, 
		       last_sync_error, local_version, is_deleted
		FROM local_secrets 
		WHERE sync_status IN ('pending', 'failed') AND is_deleted = 0
		ORDER BY updated_at ASC
		LIMIT ?`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("get pending secrets: %w", err)
	}
	defer rows.Close()

	var secrets []*LocalSecret
	for rows.Next() {
		var secret LocalSecret
		var serverID sql.NullInt64
		var lastError sql.NullString
		var metadata sql.NullString

		err := rows.Scan(
			&secret.ID, &serverID, &secret.DataType, &secret.DataJSON, &metadata,
			&secret.CreatedAt, &secret.UpdatedAt, &secret.SyncStatus, &secret.SyncAttempts,
			&lastError, &secret.LocalVersion, &secret.IsDeleted,
		)
		if err != nil {
			return nil, fmt.Errorf("scan secret: %w", err)
		}

		if serverID.Valid {
			sid := int(serverID.Int64)
			secret.ServerID = &sid
		}
		if lastError.Valid {
			secret.LastSyncError = lastError.String
		}

		secrets = append(secrets, &secret)
	}

	return secrets, nil
}

func (s *LocalStorage) MarkSecretsAsFailed(localIDs []int, errorMsg string) error {
	if len(localIDs) == 0 {
		return nil
	}

	placeholders := strings.Repeat("?,", len(localIDs))
	placeholders = placeholders[:len(placeholders)-1]

	query := fmt.Sprintf(`
		UPDATE local_secrets 
		SET sync_status = ?, 
		    sync_attempts = sync_attempts + 1,
		    last_sync_error = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = %s`, placeholders)

	args := make([]interface{}, 2+len(localIDs))
	args[0] = SyncStatusFailed
	args[1] = errorMsg
	for i, id := range localIDs {
		args[2+i] = id
	}

	_, err := s.db.Exec(query, args...)
	return err
}

func (s *LocalStorage) MarkSecretsAsSynced(localIDs map[int]int) error {
	if len(localIDs) == 0 {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for localID, serverID := range localIDs {
		_, err := tx.Exec(`
			UPDATE local_secrets 
			SET sync_status = ?, 
			    server_id = ?, 
			    sync_attempts = 0, 
			    last_sync_error = '',
			    updated_at = CURRENT_TIMESTAMP
			WHERE id = ?`, SyncStatusSynced, serverID, localID)
		if err != nil {
			return fmt.Errorf("mark as synced %d: %w", localID, err)
		}
	}

	return tx.Commit()
}

func (s *LocalStorage) GetSyncStats() (map[string]int, error) {
	stats := make(map[string]int)
	rows, err := s.db.Query(`
		SELECT sync_status, COUNT(*) 
		FROM local_secrets 
		WHERE is_deleted = 0 
		GROUP BY sync_status`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int
		rows.Scan(&status, &count)
		stats[string(SyncStatus(status))] = count
	}
	return stats, nil
}
