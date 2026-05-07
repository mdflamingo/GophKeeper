package commands

import (
	"errors"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/mdflamingo/GophKeeper/internal/client/model"
)

func validateNotEmpty(input string) error {
	if len(strings.TrimSpace(input)) == 0 {
		return errors.New("поле не может быть пустым")
	}
	return nil
}

func inputTextData() (*model.TextData, error) {
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

	return &model.TextData{Text: text}, nil

}

func inputCardData() (*model.CardData, error) {
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

	return &model.CardData{
		Number:     number,
		ExpiryDate: expiry,
		CVV:        cvv,
		HolderName: holder,
	}, nil
}

func inputCredentialsData() (*model.CredentialsData, error) {
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

	return &model.CredentialsData{
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
