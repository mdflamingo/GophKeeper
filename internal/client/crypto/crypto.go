package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/joho/godotenv"
)

func GetMasterKey() (string, error) {
	_ = godotenv.Load(".env")

	masterKey := os.Getenv("MASTER_KEY")
	if masterKey == "" {
		return "", fmt.Errorf("мастер ключ не найден в .env файле")
	}

	return masterKey, nil
}

func Encrypt(data any, masterPassword string) (json.RawMessage, error) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("json marshal: %w", err)
	}

	keyHash := sha256.Sum256([]byte(masterPassword))
	key := keyHash[:32]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}

	iv := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("iv gen: %w", err)
	}

	ciphertext := gcm.Seal(nil, iv, dataJSON, nil)

	encrypted := struct {
		CipherText string `json:"ciphertext"`
		IV         string `json:"iv"`
	}{
		CipherText: base64.StdEncoding.EncodeToString(ciphertext),
		IV:         base64.StdEncoding.EncodeToString(iv),
	}

	result, err := json.Marshal(encrypted)
	return result, err
}

func Decrypt(data json.RawMessage, masterPassword string) (map[string]interface{}, error) {
	var encrypted struct {
		CipherText string `json:"ciphertext"`
		IV         string `json:"iv"`
	}
	if err := json.Unmarshal(data, &encrypted); err != nil {
		return nil, fmt.Errorf("parse encrypted: %w", err)
	}

	keyHash := sha256.Sum256([]byte(masterPassword))
	key := keyHash[:32]

	ciphertext, err := base64.StdEncoding.DecodeString(encrypted.CipherText)
	if err != nil {
		return nil, fmt.Errorf("base64 ciphertext: %w", err)
	}
	iv, err := base64.StdEncoding.DecodeString(encrypted.IV)
	if err != nil {
		return nil, fmt.Errorf("base64 iv: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}

	plaintext, err := gcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(plaintext, &result); err != nil {
		return nil, fmt.Errorf("json unmarshal: %w", err)
	}
	return result, nil
}
