package requests

import (
	"fmt"

	"github.com/mdflamingo/GophKeeper/internal/client"
	"github.com/mdflamingo/GophKeeper/internal/model"
)

func LoginRequest(c *client.Client, username, password string) (string, error) {
	authData := model.AuthUser{
		Login:    username,
		Password: password,
	}

	var response model.AuthResponse

	resp, err := c.Client.R().
		SetBody(authData).
		SetResult(&response).
		Post("/api/user/login")

	if err != nil {
		return "", fmt.Errorf("ошибка входа: %w", err)
	}

	if resp.IsError() {
		return "", fmt.Errorf("ошибка входа: %s", string(resp.Body()))
	}

	c.SetToken(response.Token)

	fmt.Println("✅ Успешный вход!")
	return response.Token, nil
}
