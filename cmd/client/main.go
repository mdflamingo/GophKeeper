package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/mdflamingo/GophKeeper/internal/client"
	"github.com/mdflamingo/GophKeeper/internal/client/commands"
	"github.com/mdflamingo/GophKeeper/internal/client/requests"
	"github.com/mdflamingo/GophKeeper/internal/config"
	"github.com/mdflamingo/GophKeeper/internal/repository/sqlite"
)

func main() {
	conf := config.GetConfig()
	c := client.NewClient(conf.ServerAddr)

	storage, err := sqlite.New("./")
	if err != nil {
		fmt.Printf("❌ Ошибка создания хранилища: %v\n", err)
		os.Exit(1)
	}
	defer storage.Close()

	fmt.Println("🔐 Добро пожаловать в GophKeeper!")
	fmt.Println(strings.Repeat("=", 50))

	if err := initializeClient(c); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	syncCtx, syncCancel := context.WithCancel(context.Background())
	syncService := requests.New(storage, c)
	go syncService.Start(syncCtx)
	defer syncCancel()

	runMainLoop(c, storage, syncService)
}

func initializeClient(c *client.Client) error {
	tokenData, err := commands.LoadTokenFromFile()
	if err == nil && tokenData.Token != "" {
		c.SetToken(tokenData.Token)
		return nil

	}

	return showAuthMenu(c)
}

func showAuthMenu(c *client.Client) error {
	fmt.Println("\nДля работы с приложением необходимо авторизоваться")

	prompt := promptui.Select{
		Label: "Выберите действие",
		Items: []string{
			"🔑 Войти",
			"📝 Зарегистрироваться",
			"🚪 Выйти",
		},
	}

	_, action, err := prompt.Run()
	if err != nil {
		return err
	}

	switch action {
	case "🔑 Войти":
		return commands.Login(c)
	case "📝 Зарегистрироваться":
		return commands.Register(c)
	case "🚪 Выйти":
		fmt.Println("👋 До свидания!")
		os.Exit(0)
	}

	return nil
}

func runMainLoop(c *client.Client, storage *sqlite.LocalStorage, syncService *requests.SyncService) {
	for {

		prompt := promptui.Select{
			Label: "Выберите действие",
			Items: []string{
				"📝 Создать секрет",
				"📋 Показать все секреты",
				"🔍 Найти секрет по ID",
				"✏️ Редактировать секрет",
				"🚪 Выйти",
			},
			Size: 7,
		}

		_, action, err := prompt.Run()
		if err != nil {
			fmt.Printf("Ошибка: %v\n", err)
			continue
		}

		switch action {
		case "📝 Создать секрет":
			handleAction(commands.CreateSecret(c, storage))
		case "📋 Показать все секреты":
			handleAction(commands.GetSecrets(c, storage))
		case "🔍 Найти секрет по ID":
			handleAction(commands.GetSecretByID(c))
		case "✏️ Редактировать секрет":
			handleAction(commands.UpdateSecretByID(c, storage))
		case "🚪 Выйти":
			fmt.Println("👋 До свидания!")
			os.Exit(0)
		}
	}
}

func handleAction(err error) {
	if err != nil {
		fmt.Printf("❌ %v\n", err)
	}
}
