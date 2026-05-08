package requests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mdflamingo/GophKeeper/internal/client"
	"github.com/mdflamingo/GophKeeper/internal/model"
)

func CreateSecretRequest(ctx context.Context, c *client.Client, dataType model.DataType, data any) (*model.SecretCreateResponse, error) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("ошибка сериализации данных: %w", err)
	}

	request := model.SecretCreateRequest{
		DataType: dataType,
		Data:     dataJSON,
	}

	var response model.SecretCreateResponse

	resp, err := c.Client.R().
		SetContext(ctx).
		SetBody(request).
		SetResult(&response).
		Post("/api/secret")

	if err != nil {
		return nil, fmt.Errorf("ошибка отправки запроса: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("сервер вернул ошибку: %d - %s", resp.StatusCode(), string(resp.Body()))
	}

	fmt.Printf("✅ Секрет успешно создан! ID: %d\n", response.ID)
	return &response, nil
}

func GetOneSecretRequest(ctx context.Context, c *client.Client, secretID string) (*model.SecretResponse, error) {
	var response model.SecretResponse

	resp, err := c.Client.R().
		SetContext(ctx).
		SetResult(&response).
		SetPathParam("id", secretID).
		Get("/api/secret/{id}")

	if err != nil {
		return nil, fmt.Errorf("ошибка отправки запроса: %w", err)
	}

	if resp.StatusCode() == http.StatusNoContent {
		return nil, nil
	}

	if resp.IsError() {
		return nil, fmt.Errorf("сервер вернул ошибку: %d - %s", resp.StatusCode(), string(resp.Body()))
	}

	return &response, nil
}

func GetSecretListRequest(ctx context.Context, c *client.Client) (*model.SecretListResponse, error) {
	var response model.SecretListResponse

	resp, err := c.Client.R().
		SetContext(ctx).
		SetResult(&response).
		Get("/api/secret/list")

	if err != nil {
		return nil, fmt.Errorf("ошибка отправки запроса: %w", err)
	}

	if resp.StatusCode() == http.StatusNoContent {
		return nil, nil
	}

	if resp.IsError() {
		return nil, fmt.Errorf("сервер вернул ошибку: %d - %s", resp.StatusCode(), string(resp.Body()))
	}

	return &response, nil
}

func UpdateSecretRequest(ctx context.Context, c *client.Client, secretID string, dataType model.DataType, data any) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("ошибка сериализации данных: %w", err)
	}

	request := model.SecretUpdateRequest{
		DataType: dataType,
		Data:     dataJSON,
	}

	resp, err := c.Client.R().
		SetContext(ctx).
		SetBody(request).
		SetPathParam("id", secretID).
		Put("/api/secret/{id}")

	if err != nil {
		return fmt.Errorf("ошибка отправки запроса: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("сервер вернул ошибку: %d - %s", resp.StatusCode(), string(resp.Body()))
	}

	fmt.Printf("✅ Секрет #%s успешно обновлен!\n", secretID)
	return nil
}
