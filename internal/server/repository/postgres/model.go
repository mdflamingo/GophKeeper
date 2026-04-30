package postgres

import (
	"encoding/json"
	"time"
)

type DataType string

func (dt DataType) String() string {
	return string(dt)
}

const (
	TEXT        DataType = "Text"
	CARD        DataType = "Card"
	FILE        DataType = "File"
	CREDENTIALS DataType = "Credentials"
)

type UserDataDB struct {
	DataType  DataType        `json:"data_type"`
	MetaData  json.RawMessage `json:"meta_data"`
	CreatedAt time.Time       `json:"created_at"`
}

type UserDB struct {
	ID       int    `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
}
