package minio

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/mdflamingo/GophKeeper/internal/config"
	"github.com/mdflamingo/GophKeeper/internal/logger"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

type FileStorage interface {
	UploadFile(ctx context.Context, file io.Reader, bucketName, fileName string, fileSize int64) error
	DownloadFile(bucketName, fileName string) (io.ReadCloser, error)
	GetPresignedURL(bucketName, fileName string, expiry time.Duration) (string, error)
	Close() error
	Ping(ctx context.Context) error
}

func ConnectMinio(minioConf *config.Minio) (FileStorage, error) {
	logger.Log.Info("Starting initialize to MinIO",
		zap.String("endpoint", minioConf.MinioEndpoint))

	minioCreds := credentials.NewStaticV4(
		minioConf.MinioRootUser,
		minioConf.MinioRootPassword,
		"",
	)

	minioClient, err := minio.New(minioConf.MinioEndpoint, &minio.Options{
		Creds:  minioCreds,
		Secure: false,
	})
	if err != nil {
		logger.Log.Error("Failed to create MinIO client", zap.Error(err))
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	ctx := context.Background()
	_, err = minioClient.ListBuckets(ctx)
	if err != nil {
		logger.Log.Error("Failed to list MinIO buckets", zap.Error(err))
		return nil, fmt.Errorf("minio connection failed: %w", err)
	}

	logger.Log.Info("Successfully initialized MinIO storage")
	return &MinioStorage{client: minioClient}, nil
}
