package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/mdflamingo/GophKeeper/internal/client/crypto"
	clientModel "github.com/mdflamingo/GophKeeper/internal/client/model"

	apiModel "github.com/mdflamingo/GophKeeper/internal/model"
)

func safeStringValue(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok && val != "" {
		return val
	}
	return "неизвестно"
}

func safeInt64Value(data map[string]interface{}, key string) int64 {
	if val, ok := data[key].(float64); ok {
		return int64(val)
	}
	return 0
}

func parseFileMetadata(metaData json.RawMessage) (map[string]interface{}, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(metaData, &data); err != nil {
		return nil, err
	}
	return data, nil
}
func validateNotEmpty(input string) error {
	if len(strings.TrimSpace(input)) == 0 {
		return errors.New("поле не может быть пустым")
	}
	return nil
}

func inputTextData() (*clientModel.TextData, error) {
	prompt := promptui.Prompt{
		Label: "Введите текст секрета",
		Validate: func(input string) error {
			if len(input) == 0 {
				return errors.New("текст не может быть пустым")
			}
			return nil
		},
	}

	text, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	return &clientModel.TextData{Text: text}, nil

}

func inputCardData() (*clientModel.CardData, error) {
	numberPrompt := promptui.Prompt{
		Label:    "Номер карты",
		Validate: validateNotEmpty,
	}
	number, err := numberPrompt.Run()
	if err != nil {
		return nil, err
	}

	expiryPrompt := promptui.Prompt{
		Label:    "Срок действия (ММ/ГГ)",
		Validate: validateNotEmpty,
	}
	expiry, err := expiryPrompt.Run()
	if err != nil {
		return nil, err
	}

	cvvPrompt := promptui.Prompt{
		Label:    "CVV",
		Validate: validateNotEmpty,
		Mask:     '*',
	}
	cvv, err := cvvPrompt.Run()
	if err != nil {
		return nil, err
	}

	holderPrompt := promptui.Prompt{
		Label:    "Имя держателя",
		Validate: validateNotEmpty,
	}
	holder, err := holderPrompt.Run()
	if err != nil {
		return nil, err
	}

	return &clientModel.CardData{
		Number:     number,
		ExpiryDate: expiry,
		CVV:        cvv,
		HolderName: holder,
	}, nil
}

func inputCredentialsData() (*clientModel.CredentialsData, error) {
	loginPrompt := promptui.Prompt{
		Label:    "Логин",
		Validate: validateNotEmpty,
	}
	login, err := loginPrompt.Run()
	if err != nil {
		return nil, err
	}

	passPrompt := promptui.Prompt{
		Label:    "Пароль",
		Validate: validateNotEmpty,
		Mask:     '*',
	}
	password, err := passPrompt.Run()
	if err != nil {
		return nil, err
	}

	urlPrompt := promptui.Prompt{
		Label: "URL (необязательно)",
	}
	url, _ := urlPrompt.Run()

	notePrompt := promptui.Prompt{
		Label: "Заметка (необязательно)",
	}
	note, _ := notePrompt.Run()

	return &clientModel.CredentialsData{
		Login:    login,
		Password: password,
		URL:      url,
		Note:     note,
	}, nil
}

func inputFileData() (*clientModel.FileData, error) {
	filePrompt := promptui.Prompt{
		Label: "Путь к файлу",
		Validate: func(input string) error {
			if len(strings.TrimSpace(input)) == 0 {
				return errors.New("путь к файлу не может быть пустым")
			}
			return validateFileExists(input)
		},
	}

	filePath, err := filePrompt.Run()
	if err != nil {
		return nil, err
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения информации о файле: %w", err)
	}

	const maxSize = 10 * 1024 * 1024
	if fileInfo.Size() > maxSize {
		return nil, fmt.Errorf("файл слишком большой: %d байт (максимум %d)",
			fileInfo.Size(), maxSize)
	}

	return &clientModel.FileData{
		Path:     filePath,
		Filename: fileInfo.Name(),
		Size:     fileInfo.Size(),
	}, nil
}

func validateFileExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return errors.New("файл не существует")
	}
	return nil
}

func handleInput() (any, apiModel.DataType, error) {
	typePrompt := promptui.Select{
		Label: "Выберите тип секрета",
		Items: []string{
			string(apiModel.TEXT),
			string(apiModel.CARD),
			string(apiModel.CREDENTIALS),
			string(apiModel.FILE),
		},
	}
	_, typeStr, err := typePrompt.Run()
	if err != nil {
		return nil, apiModel.DataType(typeStr), fmt.Errorf("ошибка выбора типа: %w", err)
	}
	dataType := apiModel.DataType(typeStr)

	var data any

	switch dataType {
	case apiModel.TEXT:
		data, err = inputTextData()
	case apiModel.CARD:
		data, err = inputCardData()
	case apiModel.CREDENTIALS:
		data, err = inputCredentialsData()
	case apiModel.FILE:
		data, err = inputFileData()
	default:
		return nil, apiModel.DataType(typeStr), errors.New("неизвестный тип секрета")
	}

	if err != nil {
		return nil, apiModel.DataType(typeStr), fmt.Errorf("ошибка ввода данных: %w", err)
	}

	return data, dataType, nil
}

func printSecret(index int, secret apiModel.SecretResponse, masterPassword string) {
	fmt.Printf(" [%d] ID:%-4d | Тип данных: %-12s | Дата создания: %s\n",
		index, secret.ID, secret.DataType, secret.CreatedAt.Format("02.01 15:04"))

	if len(secret.MetaData) == 0 {
		fmt.Println("       📭 Нет данных")
		return
	}

	switch secret.DataType {
	case apiModel.FILE:
		fileMeta, err := parseFileMetadata(secret.MetaData)
		if err != nil {
			fmt.Printf("       ❌ Метаданные: %v\n", err)
			return
		}
		filename := safeStringValue(fileMeta, "filename")
		size := safeInt64Value(fileMeta, "size")
		url := safeStringValue(fileMeta, "download_url")

		fmt.Printf("	📄 Имя файла: %s\n", filename)
		fmt.Printf("	📏 Размер: %s\n", formatBytes(size))
		fmt.Printf("	🔗 Ссылка: %s\n", url)

	case apiModel.TEXT, apiModel.CREDENTIALS, apiModel.CARD:
		if masterPassword == "" {
			fmt.Printf("       🔒 Зашифровано (%d байт)\n", len(secret.MetaData))
			return
		}

		decrypted, err := crypto.Decrypt(secret.MetaData, masterPassword)
		if err != nil {
			fmt.Printf("       ❌ %v\n", err)
			return
		}

		preview := formatPreview(decrypted)
		fmt.Printf("	🔓 %s\n", preview)

	default:
		fmt.Printf("       ❓ %s (%d байт)\n", secret.DataType, len(secret.MetaData))
	}
}

func formatBytes(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}
	sizes := []string{"B", "KB", "MB", "GB"}
	i := 0
	for bytes >= 1024 && i < len(sizes)-1 {
		bytes /= 1024
		i++
	}
	return fmt.Sprintf("%d %s", bytes, sizes[i])
}

func formatPreview(data map[string]interface{}) string {
	if len(data) == 0 {
		return "пусто"
	}

	var parts []string
	for key, value := range data {
		strVal := fmt.Sprintf("%v", value)
		if len(strVal) > 20 {
			strVal = strVal[:17] + "..."
		}
		parts = append(parts, fmt.Sprintf("%s=%s", key, strVal))
		if len(parts) >= 2 {
			break
		}
	}

	if len(parts) == 0 {
		return "данные скрыты"
	}
	return strings.Join(parts, " | ")
}
