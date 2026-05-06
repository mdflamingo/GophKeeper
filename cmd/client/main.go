package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/mdflamingo/GophKeeper/internal/client"
	"github.com/mdflamingo/GophKeeper/internal/client/commands"
	"github.com/mdflamingo/GophKeeper/internal/config"
	"github.com/mdflamingo/GophKeeper/internal/logger"
)

func main() {
	conf := config.GetConfig()
	client := client.NewClient(conf.ServerAddr)

	for {
		actionPrompt := promptui.Select{
			Label: "Выберите действие",
			Items: []string{
				"Создать секрет",
				"Получить список секретов",
				"Получить секрет по id",
				"Редактировать секрет по id",
				"Выход",
			},
		}

		_, action, err := actionPrompt.Run()
		if err != nil {
			fmt.Printf("Ошибка: %v\n", err)
			continue
		}

		switch action {
		case "Создать секрет":
			err = commands.CreateSecret(client)
		case "Получить список секретов":
			err = commands.GetSecrets(client)
		case "Получить секрет по id":
			err = commands.GetSecretByID(client)
		case "Редактировать секрет по id":
			err = commands.UpdateSecretByID(client)
		case "Выход":
			fmt.Println("До свидания!")
			logger.Log.Info("До свидания!")
			os.Exit(0)
		}

		if err != nil {
			fmt.Printf("❌ %v\n", err)
		}

		fmt.Println("\n" + strings.Repeat("=", 50) + "\n")
	}
}
