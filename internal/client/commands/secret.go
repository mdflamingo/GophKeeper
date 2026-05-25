package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/mdflamingo/GophKeeper/internal/client"
	"github.com/mdflamingo/GophKeeper/internal/client/crypto"
	"github.com/mdflamingo/GophKeeper/internal/client/requests"
	apiModel "github.com/mdflamingo/GophKeeper/internal/model"
	"github.com/mdflamingo/GophKeeper/internal/repository/sqlite"
)

func CreateSecret(c *client.Client, storage *sqlite.LocalStorage) error {
	_, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	data, dataType, err := handleInput()
	if err != nil {
		return err
	}

	_, err = storage.SaveSecret(dataType, data)
	if err != nil {
		return fmt.Errorf("ошибка локального сохранения: %w", err)
	}

	fmt.Printf("💾 Секрет сохранен")

	return nil
}

func GetSecrets(c *client.Client, storage *sqlite.LocalStorage) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := requests.GetSecretListRequest(ctx, c)
	if err != nil {
		fmt.Printf("⚠️  Ошибка при получении списка секретов: %v\n", err)
		return nil
	}

	if response == nil || response.Count == 0 {
		fmt.Println("📭 Нет секретов")
		return nil
	}

	masterPassword, err := crypto.GetMasterKey()
	if err != nil {
		fmt.Printf("⚠️  Невозможно расшифровать: %v\n", err)
		masterPassword = ""
	}

	fmt.Println(strings.Repeat("=", 90))
	for i, secret := range response.Secrets {
		printSecret(i+1, secret, masterPassword)
		fmt.Println(strings.Repeat("-", 90))
	}
	fmt.Println(strings.Repeat("=", 90))

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
			masterPassword, err := crypto.GetMasterKey()
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

func UpdateSecretByID(c *client.Client, storage *sqlite.LocalStorage) error {
	idPrompt := promptui.Prompt{
		Label: "ID секрета для обновления",
		Validate: func(input string) error {
			if len(strings.TrimSpace(input)) == 0 {
				return errors.New("ID не может быть пустым")
			}
			return nil
		},
	}
	idStr, err := idPrompt.Run()
	if err != nil {
		return fmt.Errorf("ввод ID: %w", err)
	}

	data, dataType, err := handleInput()
	if err != nil {
		return err
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		serverID, parseErr := strconv.ParseInt(idStr, 10, 64)
		if parseErr == nil {
			localSecret, findErr := storage.GetSecretByServerID(serverID)
			if findErr == nil && localSecret != nil {
				id = int(localSecret.ID)
			} else {
				return fmt.Errorf("неверный формат ID: %w", err)
			}
		} else {
			return fmt.Errorf("неверный формат ID: %w", err)
		}
	}

	if err := storage.UpdateSecret(id, dataType, data); err != nil {
		return fmt.Errorf("ошибка обновления: %w", err)
	}
	fmt.Printf("💾 Секрет #%d обновлен локально\n", id)

	return nil
}
