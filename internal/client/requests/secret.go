package requests

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mdflamingo/GophKeeper/internal/client"
	"github.com/mdflamingo/GophKeeper/internal/client/crypto"
	clientModel "github.com/mdflamingo/GophKeeper/internal/client/model"
	apiModel "github.com/mdflamingo/GophKeeper/internal/model"
)

func CreateSecretRequest(ctx context.Context, c *client.Client, dataType apiModel.DataType, rawData any) (*apiModel.SecretCreateResponse, error) {
	masterPassword, err := crypto.GetMasterKey()
	if err != nil {
		return nil, err
	}

	encryptedData, err := crypto.Encrypt(rawData, masterPassword)
	if err != nil {
		return nil, fmt.Errorf("шифрование: %w", err)
	}

	request := apiModel.SecretCreateRequest{
		DataType: dataType,
		Data:     encryptedData,
	}

	var response apiModel.SecretCreateResponse
	resp, err := c.Client.R().
		SetContext(ctx).
		SetBody(request).
		SetResult(&response).
		Post("/api/secret")

	if err != nil || resp.IsError() {
		return nil, fmt.Errorf("API: %w", err)
	}

	fmt.Printf("✅ Секрет создан! ID: %d\n", response.ID)
	return &response, nil
}

func GetOneSecretRequest(ctx context.Context, c *client.Client, secretID string) (*apiModel.SecretResponse, error) {
	var response apiModel.SecretResponse

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

func GetSecretListRequest(ctx context.Context, c *client.Client) (*apiModel.SecretListResponse, error) {
	var response apiModel.SecretListResponse

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

func UpdateSecretRequest(ctx context.Context, c *client.Client, secretID string, dataType apiModel.DataType, rawData any) error {
	masterPassword, err := crypto.GetMasterKey()
	if err != nil {
		return err
	}

	encryptedData, err := crypto.Encrypt(rawData, masterPassword)
	if err != nil {
		return fmt.Errorf("шифрование: %w", err)
	}

	request := apiModel.SecretUpdateRequest{
		DataType: dataType,
		Data:     encryptedData,
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

func CreateFileSecretRequest(ctx context.Context, c *client.Client, fileData *clientModel.FileData) (*apiModel.SecretCreateResponse, error) {
	request := c.Client.R().
		SetContext(ctx).
		SetFile("file", fileData.Path).
		SetFormData(map[string]string{
			"data_type": string(apiModel.FILE),
			"data":      `{"filename":"` + fileData.Filename + `","size":` + fmt.Sprintf("%d", fileData.Size) + `}`,
		}).
		SetResult(&apiModel.SecretCreateResponse{})

	resp, err := request.Post("/api/secret/file")
	if err != nil {
		return nil, fmt.Errorf("ошибка отправки запроса: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("сервер вернул ошибку: %d - %s", resp.StatusCode(), string(resp.Body()))
	}

	response := resp.Result().(*apiModel.SecretCreateResponse)
	fmt.Printf("✅ Файл '%s' успешно загружен! ID: %d (размер: %d байт)\n",
		fileData.Filename, response.ID, fileData.Size)

	return response, nil
}

func UpdateFileSecretRequest(ctx context.Context, c *client.Client, secretID string, fileData *clientModel.FileData) error {
	resp, err := c.Client.R().
		SetContext(ctx).
		SetFile("file", fileData.Path).
		SetFormData(map[string]string{
			"data_type": string(apiModel.FILE),
			"data":      `{"filename":"` + fileData.Filename + `","size":` + fmt.Sprintf("%d", fileData.Size) + `}`,
		}).
		SetPathParam("id", secretID).
		Put("/api/secret/file/{id}")

	if err != nil {
		return fmt.Errorf("ошибка отправки запроса: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("сервер вернул ошибку: %d - %s", resp.StatusCode(), string(resp.Body()))
	}

	fmt.Printf("✅ Файл '%s' успешно обновлен! ID: %s\n", fileData.Filename, secretID)
	return nil
}
