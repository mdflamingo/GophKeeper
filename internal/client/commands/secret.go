package commands

import (
	"errors"
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/mdflamingo/GophKeeper/internal/client"
	"github.com/mdflamingo/GophKeeper/internal/client/requests"
	"github.com/mdflamingo/GophKeeper/internal/model"
)

func CreateSecret(c *client.Client) error {
	typePrompt := promptui.Select{
		Label: "Выберите тип секрета",
		Items: []string{
			string(model.TEXT),
			string(model.CARD),
			string(model.CREDENTIALS),
			string(model.FILE),
		},
	}
	_, typeStr, err := typePrompt.Run()
	if err != nil {
		return fmt.Errorf("ошибка выбора типа: %w", err)
	}
	dataType := model.DataType(typeStr)

	var data any

	switch dataType {
	case model.TEXT:
		data, err = inputTextData()
	case model.CARD:
		data, err = inputCardData()
	case model.CREDENTIALS:
		data, err = inputCredentialsData()
	// case model.FILE:
	// 	return createFileSecret(c)
	default:
		return errors.New("неизвестный тип секрета")
	}

	if err != nil {
		return fmt.Errorf("ошибка ввода данных: %w", err)
	}
	_, err = requests.SendCreateSecretRequest(c, dataType, data)
	return nil

}

func GetSecrets(c *client.Client) error {
	response, err := requests.GetSecretListRequest(c)
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

	secret, err := requests.GetOneSecretRequest(c, secretID)
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
	return errors.New("test error")
}
