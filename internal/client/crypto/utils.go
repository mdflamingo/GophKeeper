package crypto

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/mdflamingo/GophKeeper/internal/client"
)

func GetMasterPassword(client *client.Client) (string, error) {
	if cachedPass, ok := client.GetCachedMasterPassword(); ok {
		return cachedPass, nil
	}

	prompt := promptui.Prompt{
		Label: "🔐 Мастер-пароль",
		Mask:  '*',
	}
	pass, err := prompt.Run()
	if err != nil {
		return "", err
	}

	client.SetMasterPassword(pass)
	return pass, nil
}

func ClearMasterPasswordCache(c *client.Client) error {
	c.ClearMasterPassword()
	fmt.Println("✅ Кэш мастер-пароля очищен")
	return nil
}
