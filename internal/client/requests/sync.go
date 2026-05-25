package requests

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/mdflamingo/GophKeeper/internal/client"
	"github.com/mdflamingo/GophKeeper/internal/client/crypto"
	apiModel "github.com/mdflamingo/GophKeeper/internal/model"
	"github.com/mdflamingo/GophKeeper/internal/repository/sqlite"
)

type SyncService struct {
	storage *sqlite.LocalStorage
	client  *client.Client
	running bool
	mu      sync.RWMutex
}

func New(storage *sqlite.LocalStorage, c *client.Client) *SyncService {
	return &SyncService{
		storage: storage,
		client:  c,
	}
}

func (s *SyncService) Start(ctx context.Context) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.Stop()
			return
		case <-ticker.C:
			if err := s.SyncBatch(ctx); err != nil {
				fmt.Printf("⚠️  Ошибка синхронизации: %v\n", err)
			} else {
				fmt.Println("✅ Батч синхронизация завершена")
			}
		}
	}
}

func (s *SyncService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = false
}

func (s *SyncService) SyncBatch(ctx context.Context) error {
	secrets, err := s.storage.GetPendingSecrets(50)
	if err != nil {
		return fmt.Errorf("получение pending секретов: %w", err)
	}
	if len(secrets) == 0 {
		return nil
	}

	var localIDs []int
	for _, secret := range secrets {
		localIDs = append(localIDs, int(secret.ID))
	}

	return s.sendBatchRequest(ctx, secrets)
}

func (s *SyncService) sendBatchRequest(ctx context.Context, secrets []*sqlite.LocalSecret) error {
	batch := make([]apiModel.BatchSecret, 0, len(secrets))

	masterPassword, err := crypto.GetMasterKey()
	if err != nil {
		return fmt.Errorf("мастер-пароль для синхронизации: %w", err)
	}

	for _, secret := range secrets {
		var rawData interface{}
		if err := json.Unmarshal([]byte(secret.DataJSON), &rawData); err != nil {
			fmt.Printf("⚠️  Пропуск секрета %d: %v\n", secret.ID, err)
			continue
		}

		encryptedData, err := crypto.Encrypt(rawData, masterPassword)
		if err != nil {
			fmt.Printf("⚠️  Ошибка шифрования секрета %d: %v\n", secret.ID, err)
			continue
		}

		batchSecret := apiModel.BatchSecret{
			LocalID:      int(secret.ID),
			ServerID:     secret.ServerID,
			DataType:     apiModel.DataType(secret.DataType),
			Data:         encryptedData,
			LocalVersion: int(secret.LocalVersion),
		}
		batch = append(batch, batchSecret)
	}

	if len(batch) == 0 {
		fmt.Println("📭 Нет секретов для синхронизации")
		return nil
	}

	req := apiModel.BatchSyncRequest{Secrets: batch}

	var response apiModel.BatchSyncResponse
	resp, err := s.client.Client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&response).
		Post("/api/secret/batch")

	if err != nil {
		return fmt.Errorf("batch request failed: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("batch sync failed: %d - %s", resp.StatusCode(), string(resp.Body()))
	}

	successMap := make(map[int]int)
	failedIDs := make([]int, 0)

	for _, result := range response.Success {
		successMap[result.LocalID] = result.ServerID
	}
	for _, errItem := range response.Failed {
		failedIDs = append(failedIDs, errItem.LocalID)
	}

	if err := s.storage.MarkSecretsAsSynced(successMap); err != nil {
		return fmt.Errorf("mark success failed: %w", err)
	}
	if len(failedIDs) > 0 {
		errorMsg := fmt.Sprintf("batch sync error: %d failed", len(failedIDs))
		if err := s.storage.MarkSecretsAsFailed(failedIDs, errorMsg); err != nil {
			return fmt.Errorf("mark failed failed: %w", err)
		}
	}

	fmt.Printf("📊 Батч: %d/%d отправлено, %d создано, %d обновлено, %d ошибок\n",
		response.Stats.Processed, len(secrets), response.Stats.Created,
		response.Stats.Updated, response.Stats.Failed)

	return nil
}
