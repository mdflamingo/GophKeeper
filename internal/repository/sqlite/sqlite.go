package sqlite

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/mdflamingo/GophKeeper/internal/model"
	_ "modernc.org/sqlite"
)

type LocalStorage struct {
	db *sql.DB
}

func New(path string) (*LocalStorage, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	s := &LocalStorage{db: db}
	return s, s.initSchema()
}

func (s *LocalStorage) initSchema() error {
	schema := `
    CREATE TABLE IF NOT EXISTS secrets (
        id INTEGER PRIMARY KEY,
        server_id INTEGER,
        data_type TEXT NOT NULL,
        data BLOB NOT NULL,
        meta_data TEXT,
        created_at TEXT,
        updated_at TEXT,
        synced INTEGER DEFAULT 0
    );
    CREATE TABLE IF NOT EXISTS config (
        key TEXT PRIMARY KEY,
        value TEXT
    );
    `
	_, err := s.db.Exec(schema)
	return err
}

func (s *LocalStorage) SaveToken(token string) error {
	_, err := s.db.Exec("INSERT OR REPLACE INTO config (key, value) VALUES (?, ?)", "token", token)
	return err
}

func (s *LocalStorage) GetToken() (string, error) {
	var token string
	err := s.db.QueryRow("SELECT value FROM config WHERE key = ?", "token").Scan(&token)
	return token, err
}

func (s *LocalStorage) SaveSecret(secret model.SecretResponse, data map[string]interface{}) error {
	dataJSON, _ := json.Marshal(data)
	metaJSON, _ := json.Marshal(secret.MetaData)

	_, err := s.db.Exec(`
        INSERT OR REPLACE INTO secrets 
        (server_id, data_type, data, meta_data, created_at, synced, updated_at)
        VALUES (?, ?, ?, ?, ?, 1, ?)`,
		secret.ID, string(secret.DataType), dataJSON, metaJSON,
		secret.CreatedAt, time.Now().Format(time.RFC3339),
	)
	return err
}

func (s *LocalStorage) GetSecrets() ([]model.SecretResponse, error) {
	rows, err := s.db.Query("SELECT server_id, data_type, meta_data, created_at FROM secrets")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var secrets []model.SecretResponse
	for rows.Next() {
		var s model.SecretResponse
		var metaJSON string
		err := rows.Scan(&s.ID, &s.DataType, &metaJSON, &s.CreatedAt)
		if err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(metaJSON), &s.MetaData)
		secrets = append(secrets, s)
	}
	return secrets, nil
}

func (s *LocalStorage) Close() error {
	return s.db.Close()
}
