package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mdflamingo/GophKeeper/internal/config"
	"github.com/mdflamingo/GophKeeper/internal/logger"
	"github.com/mdflamingo/GophKeeper/internal/server/repository/postgres"
)

func NewRouter(conf *config.Config, storage postgres.Storage) *chi.Mux {
	r := chi.NewRouter()

	r.Use(logger.RequestLogger)

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		postgres.HealthCheck(w, r, storage)
	})

	return r
}
