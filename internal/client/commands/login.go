package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/manifoldco/promptui"
	"github.com/mdflamingo/GophKeeper/internal/client"
	"github.com/mdflamingo/GophKeeper/internal/client/model"
	"github.com/mdflamingo/GophKeeper/internal/client/requests"
)

func Login(c *client.Client) error {
	loginPrompt := promptui.Prompt{
		Label: "Логин",
		Validate: func(input string) error {
			if len(input) < 3 {
				return fmt.Errorf("логин должен быть не менее 3 символов")
			}
			return nil
		},
	}
	login, err := loginPrompt.Run()
	if err != nil {
		return fmt.Errorf("ввод логина отменен")
	}

	passwordPrompt := promptui.Prompt{
		Label: "Пароль",
		Mask:  '*',
		Validate: func(input string) error {
			if len(input) < 6 {
				return fmt.Errorf("пароль должен быть не менее 6 символов")
			}
			return nil
		},
	}
	password, err := passwordPrompt.Run()
	if err != nil {
		return fmt.Errorf("ввод пароля отменен")
	}

	fmt.Print("🔑 Выполняю вход... ")

	token, err := requests.LoginRequest(c, login, password)
	if err != nil {
		return err
	}

	if err := saveTokenToFile(login, token); err != nil {
		fmt.Printf("\n⚠️  Не удалось сохранить токен локально: %v\n", err)
	}

	fmt.Println("✅ Успешный вход!")
	return nil
}

func LoadTokenFromFile() (*model.TokenData, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	tokenFile := filepath.Join(homeDir, ".gophkeeper", "token.json")
	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return nil, err
	}

	var tokenData model.TokenData
	if err := json.Unmarshal(data, &tokenData); err != nil {
		return nil, err
	}

	return &tokenData, nil
}

func saveTokenToFile(login, token string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".gophkeeper")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}

	tokenData := model.TokenData{
		Token: token,
		Login: login,
	}

	data, err := json.Marshal(tokenData)
	if err != nil {
		return err
	}

	tokenFile := filepath.Join(configDir, "token.json")
	return os.WriteFile(tokenFile, data, 0600)
}
