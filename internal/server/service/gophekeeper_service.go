package service

import (
	"errors"
	"fmt"

	"github.com/mdflamingo/GophKeeper/internal/model"
	"github.com/mdflamingo/GophKeeper/internal/server/repository/postgres"
)

var (
	ErrInvalidUserID   = errors.New("invalid user ID: cannot be zero")
	ErrInvalidSecretID = errors.New("invalid secret ID: cannot be zero")
	ErrSecretNotFound  = errors.New("secret not found")
)

type GopheKeeperService struct {
	repo *postgres.DBStorage
}

func NewGopheKeeperService(repo *postgres.DBStorage) *GopheKeeperService {
	return &GopheKeeperService{repo: repo}
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
			DataType:  postgres.DataType(secret.DataType),
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
		if errors.Is(err, errors.New("no rows in result set")) {
			return nil, ErrSecretNotFound
		}
		return nil, fmt.Errorf("failed to get secret from repository: %w", err)
	}

	response := &model.SecretResponse{
		ID:        secret.ID,
		DataType:  postgres.DataType(secret.DataType),
		MetaData:  secret.MetaData,
		CreatedAt: secret.CreatedAt,
	}

	return response, nil
}
