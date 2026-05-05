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
