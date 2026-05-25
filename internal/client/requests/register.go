package requests

import (
	"fmt"

	"github.com/mdflamingo/GophKeeper/internal/client"
	"github.com/mdflamingo/GophKeeper/internal/model"
)

func RegisterRequest(c *client.Client, username, password string) (string, error) {
	authData := model.AuthUser{
		Login:    username,
		Password: password,
	}

	var response model.AuthResponse

	resp, err := c.Client.R().
		SetBody(authData).
		SetResult(&response).
		Post("/api/user/register")

	if err != nil {
		return "", fmt.Errorf("ошибка регистрации: %w", err)
	}

	if resp.IsError() {
		return "", fmt.Errorf("ошибка регистрации: %s", string(resp.Body()))
	}

	c.SetToken(response.Token)

	fmt.Println("✅ Регистрация успешна!")
	return response.Token, nil
}
