package minio

import (
	"context"
	"fmt"
	"io"

	"github.com/mdflamingo/GophKeeper/internal/logger"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
)

type MinioStorage struct {
	client *minio.Client
}

// UploadFile загружает файл в MinIO
func (s *MinioStorage) UploadFile(ctx context.Context, file io.Reader, bucketName, fileName string, fileSize int64) error {
	exists, err := s.client.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = s.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	_, err = s.client.PutObject(ctx, bucketName, fileName, file, fileSize, minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	logger.Log.Info("File uploaded successfully",
		zap.String("bucket", bucketName),
		zap.String("file", fileName),
		zap.Int64("size", fileSize))

	return nil
}

func (s *MinioStorage) DownloadFile(bucketName, fileName string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(context.Background(), bucketName, fileName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	return obj, nil
}

func (s *MinioStorage) Close() error {
	return nil
}

func (s *MinioStorage) Ping(ctx context.Context) error {
	_, err := s.client.ListBuckets(ctx)
	if err != nil {
		return fmt.Errorf("minio ping failed: %w", err)
	}
	return nil
}
