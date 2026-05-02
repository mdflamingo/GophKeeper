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

// @title           GophKeeper API
// @version         1.0
// @description     API сервиса GophKeeper для безопасного хранения данных
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@gophkeeper.com

// @license.name   Apache 2.0
// @license.url    http://www.apache.org/licenses/LICENSE-2.0.html

// @host           localhost:8080
// @BasePath       /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
// @import github.com/mdflamingo/GophKeeper/internal/model

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

	server := &http.Server{
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Для разработки HTTP
	if conf.Debug == "True" || conf.Debug == "true" || conf.Debug == "TRUE" {
		server.Addr = ":8080"
		server.Handler = r
		logger.Log.Info("Starting HTTP server (development)", zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			return fmt.Errorf("HTTP server ListenAndServe: %w", err)
		}
	} else {
		// Production с HTTPS
		certificate, privateKey := readKeys()
		server.Addr = ":443"
		server.Handler = r
		logger.Log.Info("Starting HTTPS server", zap.String("addr", server.Addr))
		if err := server.ListenAndServeTLS(certificate, privateKey); err != http.ErrServerClosed {
			return fmt.Errorf("HTTPS server ListenAndServeTLS: %w", err)
		}
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

	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		logger.Log.Fatal("certificate file does not exist",
			zap.String("path", certPath),
			zap.Error(err))
	}

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		logger.Log.Fatal("private key file does not exist",
			zap.String("path", keyPath),
			zap.Error(err))
	}

	return certPath, keyPath
}
