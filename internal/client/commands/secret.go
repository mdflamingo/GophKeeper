package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/mdflamingo/GophKeeper/internal/client"
	"github.com/mdflamingo/GophKeeper/internal/client/crypto"
	clientModel "github.com/mdflamingo/GophKeeper/internal/client/model"
	"github.com/mdflamingo/GophKeeper/internal/client/requests"
	apiModel "github.com/mdflamingo/GophKeeper/internal/model"
)

func CreateSecret(c *client.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	data, dataType, err := handleInput()
	if err != nil {
		return err
	}

	switch dataType {
	case apiModel.FILE:
		fileData := data.(*clientModel.FileData)
		_, err = requests.CreateFileSecretRequest(ctx, c, fileData)
	default:
		_, err = requests.CreateSecretRequest(ctx, c, dataType, data)
	}

	return err
}

func GetSecrets(c *client.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := requests.GetSecretListRequest(ctx, c)
	if err != nil {
		return fmt.Errorf("список секретов: %w", err)
	}

	if response == nil || response.Count == 0 {
		fmt.Println("📭 У вас нет сохраненных секретов")
		return nil
	}

	fmt.Printf("\n📋 Всего секретов: %d\n", response.Count)
	fmt.Println(strings.Repeat("=", 70))

	for i, secret := range response.Secrets {
		fmt.Printf("[%d] ID: %d | Тип: %-12s | Создан: %s\n",
			i+1, secret.ID, secret.DataType, secret.CreatedAt.Format("02.01 15:04"))

		if len(secret.MetaData) > 0 {
			if secret.DataType == apiModel.FILE {
				fileMeta, err := parseFileMetadata(secret.MetaData)
				if err != nil {
					fmt.Printf("     📎 Файл (ошибка метаданных)\n")
				} else {
					filename := safeStringValue(fileMeta, "filename")
					size := safeInt64Value(fileMeta, "size")
					fmt.Printf("     📎 %s (%d байт)\n", filename, size)
				}
			} else {
				fmt.Printf("     🔐 Зашифровано (введите ID для просмотра)\n")
			}
		}
		fmt.Println(strings.Repeat("-", 70))
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
		return fmt.Errorf("ввод ID: %w", err)
	}

	secret, err := requests.GetOneSecretRequest(ctx, c, secretID)
	if err != nil {
		return fmt.Errorf("получение секрета: %w", err)
	}
	if secret == nil {
		fmt.Printf("❌ Секрет #%s не найден\n", secretID)
		return nil
	}

	fmt.Printf("\n🔐 Секрет #%d\n", secret.ID)
	fmt.Printf("📋 Тип: %s\n", secret.DataType)
	fmt.Printf("📅 Создан: %s\n", secret.CreatedAt.Format("02.01.2006 15:04:05"))
	fmt.Println(strings.Repeat("=", 50))

	if len(secret.MetaData) > 0 {
		if secret.DataType == apiModel.FILE {
			fileMeta, err := parseFileMetadata(secret.MetaData)
			if err != nil {
				fmt.Printf("❌ Ошибка разбора метаданных файла: %v\n", err)
			} else {
				fmt.Println("📎 Файловая информация:")
				fmt.Printf("   📄 Имя файла: %s\n", safeStringValue(fileMeta, "filename"))
				fmt.Printf("   📏 Размер: %d байт\n", safeInt64Value(fileMeta, "size"))
				fmt.Printf("   🔗 Ссылка: %s\n", safeStringValue(fileMeta, "download_url"))
				fmt.Printf("   ⏰ Истекает: %s\n", safeStringValue(fileMeta, "url_expires_at"))
				fmt.Println("\n💡 Скопируйте ссылку в браузер!")
			}
		} else {
			masterPassword, err := crypto.GetMasterPassword(c)
			if err != nil {
				fmt.Println("⚠️  Ввод мастер-пароля отменен")
				return nil
			}

			decrypted, err := crypto.Decrypt(secret.MetaData, masterPassword)
			if err != nil {
				fmt.Printf("❌ Ошибка расшифровки: %v\n", err)
				fmt.Println("💡 Проверьте мастер-пароль!")
			} else {
				pretty, _ := json.MarshalIndent(decrypted, "", "  ")
				fmt.Printf("🔓 Расшифрованные данные:\n%s\n", string(pretty))
			}
		}
	} else {
		fmt.Println("📭 Нет данных")
	}
	return nil
}

func UpdateSecretByID(c *client.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	idPrompt := promptui.Prompt{
		Label: "ID секрета для обновления",
		Validate: func(input string) error {
			if len(strings.TrimSpace(input)) == 0 {
				return errors.New("ID не может быть пустым")
			}
			return nil
		},
	}
	secretID, err := idPrompt.Run()
	if err != nil {
		return fmt.Errorf("ввод ID: %w", err)
	}

	data, dataType, err := handleInput()
	if err != nil {
		return err
	}

	switch dataType {
	case apiModel.FILE:
		fileData := data.(*clientModel.FileData)
		err = requests.UpdateFileSecretRequest(ctx, c, secretID, fileData)
	default:
		err = requests.UpdateSecretRequest(ctx, c, secretID, dataType, data)
	}
	return err
}
