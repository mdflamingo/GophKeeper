package model

import (
	"encoding/json"
	"time"

	"github.com/mdflamingo/GophKeeper/internal/server/repository/postgres"
)

// UserDataResponse представляет данные секрета в ответе API
type SecretResponse struct {
	ID        int               `json:"id" example:"5" description:"ID секрета"`
	DataType  postgres.DataType `json:"data_type" enums:"Text,Card,File,Credentials" example:"Credentials" description:"Тип секрета"`
	MetaData  json.RawMessage   `json:"meta_data" swaggertype:"object" description:"Метаданные секрета (зависит от типа)"`
	CreatedAt time.Time         `json:"created_at" example:"2024-01-01T12:00:00Z" description:"Дата создания"`
}

// UserDataListResponse представляет ответ со списком секретов
type SecretListResponse struct {
	Secrets []SecretResponse `json:"secrets" description:"Список секретов пользователя"`
	Count   int              `json:"count" example:"5" description:"Количество секретов"`
}
