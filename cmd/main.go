package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/h3ll0kitt1/avitotest/internal/config"
	"github.com/h3ll0kitt1/avitotest/internal/file"
	"github.com/h3ll0kitt1/avitotest/internal/logger"
	"github.com/h3ll0kitt1/avitotest/internal/storage"
	"github.com/h3ll0kitt1/avitotest/internal/storage/sql"
	"github.com/h3ll0kitt1/avitotest/internal/validator"
)

type application struct {
	storage   storage.Storage
	router    *chi.Mux
	file      file.File
	logger    *zap.SugaredLogger
	validator validator.Validator
}

// @title Avito Test API
// @version 1.0
// @description API для управления сегментами пользователей

// @host localhost:8000
// @BasePath /
func main() {

	cfg := config.NewConfig()

	r := chi.NewRouter()
	l := logger.NewLogger()
	f := file.NewCSV(cfg.Filename)
	v := validator.New()

	s, err := sql.NewStorage(cfg.Database)
	if err != nil {
		log.Fatalf("Error %s open database", err)
	}

	app := &application{
		storage:   s,
		router:    r,
		file:      f,
		logger:    l,
		validator: v,
	}
	app.setRouters()

	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: app.router,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Error %s launching server", err)
	}

}
