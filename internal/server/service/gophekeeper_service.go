package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sync"
	"time"

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
func (s *GopheKeeperService) GetSecrets(userID int, bucketName string) (*model.SecretListResponse, error) {
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

	var wg sync.WaitGroup
	var mu sync.Mutex

	errorsChan := make(chan error, len(secrets))

	for i, secret := range secrets {
		wg.Add(1)
		go func(index int, sec postgres.SecretDB) {
			defer wg.Done()

			response := model.SecretResponse{
				ID:        sec.ID,
				DataType:  model.DataType(sec.DataType),
				MetaData:  sec.MetaData,
				CreatedAt: sec.CreatedAt,
			}

			fileName := getUniqueFileName(sec.MetaData)
			if fileName != "" {
				fileResponse, err := processFile(s.minio, sec, fileName, bucketName)
				if err != nil {
					errorsChan <- fmt.Errorf("failed to process file for secret %d: %w", sec.ID, err)
				} else {
					response = *fileResponse
				}
			}

			mu.Lock()
			result[index] = response
			mu.Unlock()
		}(i, secret)
	}

	wg.Wait()
	close(errorsChan)

	response := &model.SecretListResponse{
		Secrets: result,
		Count:   len(result),
	}

	return response, nil
}

// GetOneSecret возвращает секрет в виде response модели
func (s *GopheKeeperService) GetOneSecret(userID, secretID int, bucketName string) (*model.SecretResponse, error) {
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

	fileName := getUniqueFileName(secret.MetaData)
	if fileName != "" {
		response, err := processFile(s.minio, secret, fileName, bucketName)
		if err != nil {
			return nil, fmt.Errorf("failed to get secret from repository: %w", err)
		}
		return response, nil

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
func (s *GopheKeeperService) SaveOneSecret(secret model.SecretCreateRequest, userID int) (int, error) {
	if userID == 0 {
		return 0, ErrInvalidUserID
	}
	secretID, err := s.repo.Save(secret, userID)
	if err != nil {
		return secretID, ErrSecretSave
	}

	return secretID, nil
}

func (s *GopheKeeperService) SaveFile(
	ctx context.Context,
	file io.Reader,
	fileName string,
	userID int,
	bucketName string,
	fileSize int64,
	inputSecret model.SecretCreateRequest) (int, error) {
	if userID == 0 {
		return 0, ErrInvalidUserID
	}

	uniqueFileName := generateUniqueFileName(fileName)

	err := s.minio.UploadFile(ctx, file, bucketName, uniqueFileName, fileSize)
	if err != nil {
		return 0, fmt.Errorf("failed to upload file to minio: %w", err)
	}

	if inputSecret.Data == nil || len(inputSecret.Data) == 0 {
		newFields := map[string]interface{}{
			"original_file_name": fileName,
			"unique_file_name":   uniqueFileName,
		}
		metaBytes, err := enrichMetadata(nil, newFields)
		if err != nil {
			// _ = s.minio.DeleteFile(ctx, bucketName, fileName)
			return 0, fmt.Errorf("failed to create metadata: %w", err)
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
			return 0, fmt.Errorf("failed to enrich metadata: %w", err)
		}
		inputSecret.Data = enrichedMeta
	}

	secretID, err := s.repo.Save(inputSecret, userID)
	if err != nil {
		// _ = s.minio.DeleteFile(ctx, bucketName, fileName)
		return 0, ErrSecretSave
	}

	return secretID, nil
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

func generateUniqueFileName(originalName string) string {
	ext := filepath.Ext(originalName)
	nameWithoutExt := originalName[0 : len(originalName)-len(ext)]

	return fmt.Sprintf("%s_%s%s",
		uuid.New().String(),
		nameWithoutExt,
		ext)

}

func getUniqueFileName(metadata []byte) string {
	var data map[string]interface{}
	if err := json.Unmarshal(metadata, &data); err != nil {
		return ""
	}

	value, exists := data["unique_file_name"]
	if !exists {
		return ""
	}

	strValue, ok := value.(string)
	if !ok {
		return ""
	}

	return strValue
}

func processFile(minio minio.FileStorage, secret postgres.SecretDB, fileName string, bucketName string) (*model.SecretResponse, error) {
	modifiedMetadata := make(map[string]interface{})
	if err := json.Unmarshal(secret.MetaData, &modifiedMetadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	presignedURL, err := minio.GetPresignedURL(bucketName, fileName, 15*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	delete(modifiedMetadata, "unique_file_name")
	modifiedMetadata["download_url"] = presignedURL
	modifiedMetadata["url_expires_at"] = time.Now().Add(15 * time.Minute).Format(time.RFC3339)

	metadataJSON, err := json.Marshal(modifiedMetadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	response := &model.SecretResponse{
		ID:        secret.ID,
		DataType:  model.DataType(secret.DataType),
		MetaData:  metadataJSON,
		CreatedAt: secret.CreatedAt,
	}

	return response, nil
}

// UpdateSecret обновляет существующий секрет
func (s *GopheKeeperService) UpdateSecret(secretID, userID int, updateReq model.SecretUpdateRequest) error {
	err := s.repo.Update(secretID, userID, updateReq)
	if err != nil {
		return fmt.Errorf("failed to update secret: %w", err)
	}

	if err != nil {
		return fmt.Errorf("failed to get updated secret: %w", err)
	}
	return nil
}
