package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	clientModel "github.com/mdflamingo/GophKeeper/internal/client/model"
	"golang.org/x/crypto/pbkdf2"
)

const (
	SaltSize   = 16
	KeySize    = 32
	NonceSize  = 12
	Iterations = 100000
)

type Encryptor struct {
	masterKey []byte
}

func NewEncryptor(masterPassword string) *Encryptor {
	salt := make([]byte, SaltSize)
	copy(salt, []byte("gophkeeper_salt_"))

	key := pbkdf2.Key(
		[]byte(masterPassword),
		salt,
		Iterations,
		KeySize,
		sha256.New,
	)

	return &Encryptor{
		masterKey: key,
	}
}

func (e *Encryptor) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.masterKey)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания шифра: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("ошибка генерации nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

func (e *Encryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.masterKey)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания шифра: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("некорректные зашифрованные данные")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка расшифровки: %w", err)
	}

	return plaintext, nil
}

func (e *Encryptor) EncryptStructured(data any) (*clientModel.EncryptedData, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("ошибка сериализации: %w", err)
	}

	encrypted, err := e.Encrypt(jsonData)
	if err != nil {
		return nil, fmt.Errorf("ошибка шифрования: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(encrypted)

	metadata := e.createMetadata(data)

	return &clientModel.EncryptedData{
		EncryptedContent: encoded,
		MetaData:         metadata,
	}, nil
}

func (e *Encryptor) DecryptStructured(encryptedData *clientModel.EncryptedData, target any) error {
	encrypted, err := base64.StdEncoding.DecodeString(encryptedData.EncryptedContent)
	if err != nil {
		return fmt.Errorf("ошибка декодирования: %w", err)
	}

	jsonData, err := e.Decrypt(encrypted)
	if err != nil {
		return fmt.Errorf("ошибка расшифровки: %w", err)
	}

	if err := json.Unmarshal(jsonData, target); err != nil {
		return fmt.Errorf("ошибка десериализации: %w", err)
	}

	return nil
}

func (e *Encryptor) createMetadata(data any) map[string]any {
	metadata := make(map[string]any)

	switch v := data.(type) {
	case *clientModel.TextData:
		metadata["type"] = "text"
		metadata["length"] = len(v.Text)
		metadata["preview"] = truncateString(v.Text, 50)

	case *clientModel.CardData:
		metadata["type"] = "card"
		if len(v.Number) >= 4 {
			metadata["last_four"] = v.Number[len(v.Number)-4:]
		}
		metadata["holder"] = v.HolderName

	case *clientModel.CredentialsData:
		metadata["type"] = "credentials"
		metadata["login"] = v.Login
		if v.URL != "" {
			metadata["url"] = v.URL
		}
	}

	return metadata
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
