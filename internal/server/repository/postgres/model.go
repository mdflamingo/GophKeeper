package postgres

import (
	"encoding/json"
	"time"

	"github.com/mdflamingo/GophKeeper/internal/model"
)

type SecretDB struct {
	ID        int             `json:"id"`
	DataType  model.DataType  `json:"data_type"`
	MetaData  json.RawMessage `json:"meta_data"`
	CreatedAt time.Time       `json:"created_at"`
}

type UserDB struct {
	ID       int    `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
}
