package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/mdflamingo/GophKeeper/internal/model"
	"github.com/mdflamingo/GophKeeper/internal/server/repository/minio"
	"github.com/mdflamingo/GophKeeper/internal/server/repository/postgres"
)

var (
	ErrInvalidUserID   = errors.New("invalid user ID: cannot be zero")
	ErrInvalidSecretID = errors.New("invalid secret ID: cannot be zero")
	ErrSecretNotFound  = errors.New("secret not found")
	ErrSecretSave      = errors.New("secret not saved")
	ErrFileSave        = errors.New("file not sent to minio")
)

type GopheKeeperService struct {
	repo  *postgres.DBStorage
	minio minio.FileStorage
}

func NewGopheKeeperService(repo *postgres.DBStorage, minio minio.FileStorage) *GopheKeeperService {
	return &GopheKeeperService{repo: repo, minio: minio}
}

// GetSecrets возвращает список секретов в виде response моделей
func (s *GopheKeeperService) GetSecrets(userID int) (*model.SecretListResponse, error) {
	if userID == 0 {
		return nil, ErrInvalidUserID
	}
	secrets, err := s.repo.GetList(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get secrets from repository: %w", err)
	}

	if len(secrets) == 0 {
		return &model.SecretListResponse{
			Secrets: []model.SecretResponse{},
			Count:   0,
		}, nil
	}

	result := make([]model.SecretResponse, len(secrets))
	for i, secret := range secrets {
		result[i] = model.SecretResponse{
			ID:        secret.ID,
			DataType:  model.DataType(secret.DataType),
			MetaData:  secret.MetaData,
			CreatedAt: secret.CreatedAt,
		}
	}

	response := &model.SecretListResponse{
		Secrets: result,
		Count:   len(secrets),
	}

	return response, nil
}

// GetOneSecret возвращает секрет в виде response модели
func (s *GopheKeeperService) GetOneSecret(userID, secretID int) (*model.SecretResponse, error) {
	if userID == 0 {
		return nil, ErrInvalidUserID
	}
	if secretID == 0 {
		return nil, ErrInvalidSecretID
	}

	secret, err := s.repo.Get(userID, secretID)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return nil, ErrSecretNotFound
		}
		return nil, fmt.Errorf("failed to get secret from repository: %w", err)
	}

	response := &model.SecretResponse{
		ID:        secret.ID,
		DataType:  model.DataType(secret.DataType),
		MetaData:  secret.MetaData,
		CreatedAt: secret.CreatedAt,
	}

	return response, nil
}

// SaveOneSecret сохраняет секрет
func (s *GopheKeeperService) SaveOneSecret(secret model.SecretCreateRequest, userID int) error {
	if userID == 0 {
		return ErrInvalidUserID
	}
	err := s.repo.Save(secret, userID)
	if err != nil {
		return ErrSecretSave
	}

	return nil
}

func (s *GopheKeeperService) SaveFile(
	ctx context.Context,
	file io.Reader,
	fileName string,
	userID int,
	bucketName string,
	fileSize int64,
	inputSecret model.SecretCreateRequest) error {
	if userID == 0 {
		return ErrInvalidUserID
	}

	uniqueFileName := generateUniqueFileName(fileName)

	err := s.minio.UploadFile(ctx, file, bucketName, uniqueFileName, fileSize)
	if err != nil {
		return fmt.Errorf("failed to upload file to minio: %w", err)
	}

	if inputSecret.Data == nil || len(inputSecret.Data) == 0 {
		newFields := map[string]interface{}{
			"original_file_name": fileName,
			"unique_file_name":   uniqueFileName,
		}
		metaBytes, err := enrichMetadata(nil, newFields)
		if err != nil {
			// _ = s.minio.DeleteFile(ctx, bucketName, fileName)
			return fmt.Errorf("failed to create metadata: %w", err)
		}
		inputSecret.Data = metaBytes
	} else {
		newFields := map[string]interface{}{
			"original_file_name": fileName,
			"unique_file_name":   uniqueFileName,
		}
		enrichedMeta, err := enrichMetadata(inputSecret.Data, newFields)
		if err != nil {
			// _ = s.minio.DeleteFile(ctx, bucketName, fileName)
			return fmt.Errorf("failed to enrich metadata: %w", err)
		}
		inputSecret.Data = enrichedMeta
	}

	err = s.repo.Save(inputSecret, userID)
	if err != nil {
		// _ = s.minio.DeleteFile(ctx, bucketName, fileName)
		return ErrSecretSave
	}

	return nil
}

func generateUniqueFileName(originalName string) string {
	ext := filepath.Ext(originalName)
	nameWithoutExt := originalName[0 : len(originalName)-len(ext)]

	return fmt.Sprintf("%s_%s%s",
		uuid.New().String(),
		nameWithoutExt,
		ext)

}

func enrichMetadata(existingJSON json.RawMessage, newFields map[string]interface{}) (json.RawMessage, error) {
	var metadataMap map[string]interface{}

	if existingJSON != nil && len(existingJSON) > 0 {
		if err := json.Unmarshal(existingJSON, &metadataMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal existing metadata: %w", err)
		}
	} else {
		metadataMap = make(map[string]interface{})
	}

	for key, value := range newFields {
		metadataMap[key] = value
	}

	return json.Marshal(metadataMap)
}
