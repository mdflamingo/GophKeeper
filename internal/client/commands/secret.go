package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/mdflamingo/GophKeeper/internal/client"
	"github.com/mdflamingo/GophKeeper/internal/client/requests"
)

var ErrCreateSecret = errors.New("секрет не создан")

func CreateSecret(c *client.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	data, dataType, err := handleInput()
	if err != nil {
		return err
	}
	_, err = requests.CreateSecretRequest(ctx, c, dataType, data)
	if err != nil {
		return err
	}
	return nil

}

func GetSecrets(c *client.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := requests.GetSecretListRequest(ctx, c)
	if err != nil {
		return fmt.Errorf("ошибка получения списка секретов: %w", err)
	}

	if response == nil || response.Count == 0 {
		fmt.Println("📭 У вас пока нет сохраненных секретов")
		return nil
	}

	fmt.Printf("\n📋 Найдено секретов: %d\n", response.Count)
	fmt.Println(strings.Repeat("=", 60))

	for _, secret := range response.Secrets {
		fmt.Printf("ID: %d | Тип: %s | Создан: %s\n",
			secret.ID,
			secret.DataType,
			secret.CreatedAt.Format("02.01.2006 15:04"))

		if len(secret.MetaData) > 0 {
			fmt.Printf("Данные: %s\n", string(secret.MetaData))
		}
		fmt.Println(strings.Repeat("-", 60))
	}

	return nil
}

func GetSecretByID(c *client.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	idPrompt := promptui.Prompt{
		Label: "Введите ID секрета",
		Validate: func(input string) error {
			if len(strings.TrimSpace(input)) == 0 {
				return errors.New("ID не может быть пустым")
			}
			return nil
		},
	}

	secretID, err := idPrompt.Run()
	if err != nil {
		return fmt.Errorf("ошибка ввода ID: %w", err)
	}

	secret, err := requests.GetOneSecretRequest(ctx, c, secretID)
	if err != nil {
		return fmt.Errorf("ошибка получения секрета: %w", err)
	}

	if secret == nil {
		fmt.Printf("🔍 Секрет с ID %s не найден\n", secretID)
		return nil
	}

	fmt.Printf("\n🔐 Секрет #%d\n", secret.ID)
	fmt.Println(strings.Repeat("=", 40))
	fmt.Printf("Тип: %s\n", secret.DataType)
	fmt.Printf("Создан: %s\n", secret.CreatedAt.Format("02.01.2006 15:04:05"))
	if len(secret.MetaData) > 0 {
		fmt.Printf("Данные: %s\n", string(secret.MetaData))
	}

	return nil
}

func UpdateSecretByID(client *client.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	idPrompt := promptui.Prompt{
		Label: "Введите ID секрета",
		Validate: func(input string) error {
			if len(strings.TrimSpace(input)) == 0 {
				return errors.New("ID не может быть пустым")
			}
			return nil
		},
	}

	secretID, err := idPrompt.Run()
	if err != nil {
		return fmt.Errorf("ошибка ввода ID: %w", err)
	}

	data, dataType, err := handleInput()
	if err != nil {
		return err
	}

	err = requests.UpdateSecretRequest(ctx, client, secretID, dataType, data)
	return nil
}
