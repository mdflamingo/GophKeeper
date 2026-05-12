package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
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
