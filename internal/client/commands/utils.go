package commands

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
	clientModel "github.com/mdflamingo/GophKeeper/internal/client/model"

	apiModel "github.com/mdflamingo/GophKeeper/internal/model"
)

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
	// case model.FILE:
	// 	return createFileSecret(c)
	default:
		return nil, apiModel.DataType(typeStr), errors.New("неизвестный тип секрета")
	}

	if err != nil {
		return nil, apiModel.DataType(typeStr), fmt.Errorf("ошибка ввода данных: %w", err)
	}

	return data, dataType, nil
}
