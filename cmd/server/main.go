package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"time"

	"github.com/mdflamingo/GophKeeper/internal/config"
	"github.com/mdflamingo/GophKeeper/internal/logger"
	"github.com/mdflamingo/GophKeeper/internal/server/handler"
	"github.com/mdflamingo/GophKeeper/internal/server/repository/postgres"
	"go.uber.org/zap"
)

func main() {
	conf := config.GetConfig()
	if err := run(conf); err != nil {
		log.Fatal(err)
	}
	logger.Log.Info("Server shutdown gracefully")
}

func run(conf *config.Config) error {
	if err := logger.InitLogger(conf.LogLevel); err != nil {
		return err
	}

	storage, errStorage := postgres.ConnectPG(&conf.DataBaseDSN)
	if errStorage != nil {
		logger.Log.Fatal("Failed to create storage", zap.Error(errStorage))
	}

	r := handler.NewRouter(conf, storage)

	// Cервер с базовыми настройками таймаутов
	server := &http.Server{
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	certificate, privateKey := readKeys()
	server.Addr = ":443"
	server.Handler = r

	logger.Log.Info("Starting HTTPS server", zap.String("addr", server.Addr))
	if err := server.ListenAndServeTLS(certificate, privateKey); err != http.ErrServerClosed {
		return fmt.Errorf("HTTPS server ListenAndServeTLS: %w", err)
	}
	return nil
}

// readKeys Загружает сертификат и приватный ключ из файлов ~/cert.pem и ~/private.pem
func readKeys() (string, string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Log.Fatal("cannot get user home directory", zap.Error(err))
	}

	certPath := filepath.Join(homeDir, "cert.pem")
	keyPath := filepath.Join(homeDir, "private.pem")

	certificateBytes, err := os.ReadFile(certPath)
	if err != nil {
		logger.Log.Fatal("cannot read certificate file",
			zap.String("path", certPath),
			zap.Error(err))
	}

	privateKeyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		logger.Log.Fatal("cannot read private key file",
			zap.String("path", keyPath),
			zap.Error(err))
	}

	return string(certificateBytes), string(privateKeyBytes)
}
