package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/mdflamingo/GophKeeper/api/swagger"
	"github.com/mdflamingo/GophKeeper/internal/config"
	"github.com/mdflamingo/GophKeeper/internal/logger"
	"github.com/mdflamingo/GophKeeper/internal/server/repository/postgres"
	"github.com/mdflamingo/GophKeeper/internal/server/service"
)

// NewRouter godoc
// @title           GophKeeper API
// @version         1.0
// @description     API для безопасного хранения и управления данными
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
func NewRouter(conf *config.Config, storage postgres.Storage) *chi.Mux {
	r := chi.NewRouter()

	userService := service.NewUserService(storage.(*postgres.DBStorage))
	gophekeeperService := service.NewGopheKeeperService(storage.(*postgres.DBStorage))

	r.Use(logger.RequestLogger)

	r.Group(func(r chi.Router) {
		r.Get("/api/ping", func(w http.ResponseWriter, r *http.Request) {
			DBHealthCheck(w, r, storage)
		})
		r.Post("/api/user/register", func(w http.ResponseWriter, r *http.Request) {
			RegisterHandler(w, r, userService, conf.SecretKey)
		})
		r.Post("/api/user/login", func(w http.ResponseWriter, r *http.Request) {
			LoginHandler(w, r, userService, conf.SecretKey)
		})
	})

	// Защищенные маршруты (требуют аутентификации)
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware(conf.SecretKey))

		r.Get("/api/secret/list", func(w http.ResponseWriter, r *http.Request) {
			GetSecretListHandler(w, r, gophekeeperService)
		})
		r.Get("/api/secret/{id}", func(w http.ResponseWriter, r *http.Request) {
			GetSecretHandler(w, r, gophekeeperService)
		})

		r.Post("/api/secret", func(w http.ResponseWriter, r *http.Request) {
			SaveSecretHandler(w, r, gophekeeperService)
		})
	})

	// Swagger документация
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
		httpSwagger.UIConfig(map[string]string{
			"persistAuthorization": "true",
		}),
	))

	// r.Get("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
	// 	doc, _ := swag.ReadDoc()
	// 	w.Header().Set("Content-Type", "application/json")
	// 	w.Write([]byte(doc))
	// })
	return r
}
