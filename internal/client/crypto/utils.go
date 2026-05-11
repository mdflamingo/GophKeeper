package crypto

import "github.com/manifoldco/promptui"

var masterPasswordCache string

func GetMasterPassword() (string, error) {
	if masterPasswordCache == "" {
		prompt := promptui.Prompt{
			Label: "🔐 Мастер-пароль",
			Mask:  '*',
		}
		masterPasswordCache, _ = prompt.Run()
	}
	return masterPasswordCache, nil
}
