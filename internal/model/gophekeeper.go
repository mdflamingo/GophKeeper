package model

import (
	"encoding/json"
	"time"
)

type DataType string

func (dt DataType) String() string {
	return string(dt)
}

const (
	TEXT        DataType = "TEXT"
	CARD        DataType = "CARD"
	FILE        DataType = "FILE"
	CREDENTIALS DataType = "CREDENTIALS"
)

// SecretResponse представляет данные секрета в ответе API
type SecretResponse struct {
	ID        int             `json:"id" example:"5" description:"ID секрета"`
	DataType  DataType        `json:"data_type" example:"CREDENTIALS" description:"Тип секрета"`
	MetaData  json.RawMessage `json:"meta_data" swaggertype:"object" description:"Метаданные секрета"`
	CreatedAt time.Time       `json:"created_at" example:"2024-01-01T12:00:00Z" description:"Дата создания"`
}

// SecretListResponse представляет ответ со списком секретов
type SecretListResponse struct {
	Secrets []SecretResponse `json:"secrets" description:"Список секретов пользователя"`
	Count   int              `json:"count" example:"5" description:"Количество секретов"`
}

// SecretCreateRequest тело для создания секрета
type SecretCreateRequest struct {
	DataType DataType        `json:"data_type" example:"CREDENTIALS" description:"Тип секрета (обязательное поле)"`
	Data     json.RawMessage `json:"data" swaggertype:"object" description:"Данные секрета в JSON формате (обязательное поле)"`
}

// SecretUpdateRequest представляет запрос на обновление секрета
type SecretUpdateRequest struct {
	DataType DataType        `json:"data_type" example:"CREDENTIALS" description:"Тип секрета (обязательное поле)"`
	Data     json.RawMessage `json:"data" swaggertype:"object" description:"Данные секрета в JSON формате (обязательное поле)"`
}

// SecretUpdateRequest представляет ответ с ID созданного секрета
type SecretCreateResponse struct {
	ID int `json:"id" example:"5" description:"ID сохраенного секрета секрета"`
}

// BatchSyncRequest запрос на батчевую синхронизацию
type BatchSyncRequest struct {
	Secrets []BatchSecret `json:"secrets" validate:"dive"`
}

// BatchSecret один секрет в батче
type BatchSecret struct {
	LocalID      int             `json:"local_id" validate:"required"`
	ServerID     *int            `json:"server_id,omitempty"`
	DataType     DataType        `json:"data_type" validate:"required,oneof=TEXT FILE CARD CREDENTIALS"`
	Data         json.RawMessage `json:"data" validate:"required"`
	LocalVersion int             `json:"local_version" validate:"required"`
}

// BatchSyncResponse ответ на батчевую синхронизацию
type BatchSyncResponse struct {
	Success []BatchSyncResult `json:"success"`
	Failed  []BatchSyncError  `json:"failed"`
	Stats   BatchStats        `json:"stats"`
}

type BatchSyncResult struct {
	LocalID  int `json:"local_id"`
	ServerID int `json:"server_id"`
}

type BatchSyncError struct {
	LocalID  int    `json:"local_id"`
	Error    string `json:"error"`
	ServerID *int   `json:"server_id,omitempty"`
}

type BatchStats struct {
	Processed int `json:"processed"`
	Created   int `json:"created"`
	Updated   int `json:"updated"`
	Failed    int `json:"failed"`
}
