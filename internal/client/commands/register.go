package commands

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/mdflamingo/GophKeeper/internal/client"
	"github.com/mdflamingo/GophKeeper/internal/client/requests"
)

func Register(c *client.Client) error {
	fmt.Println("📝 Регистрация нового пользователя")
	fmt.Println(strings.Repeat("-", 40))

	loginPrompt := promptui.Prompt{
		Label: "Придумайте логин",
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
		Label: "Придумайте пароль",
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

	confirmPrompt := promptui.Prompt{
		Label: "Подтвердите пароль",
		Mask:  '*',
		Validate: func(input string) error {
			if input != password {
				return fmt.Errorf("пароли не совпадают")
			}
			return nil
		},
	}
	_, err = confirmPrompt.Run()
	if err != nil {
		return fmt.Errorf("подтверждение пароля отменено")
	}

	fmt.Print("📝 Выполняю регистрацию... ")

	token, err := requests.RegisterRequest(c, login, password)
	if err != nil {
		return err
	}

	if err := saveTokenToFile(login, token); err != nil {
		fmt.Printf("\n⚠️  Не удалось сохранить токен локально: %v\n", err)
	}

	fmt.Println("✅ Регистрация успешна!")
	return nil
}
