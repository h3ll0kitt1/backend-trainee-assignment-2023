package main

import (
	"log"
	"net/http"
	"time"

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

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Error %s load configuration", err)
	}

	r := chi.NewRouter()
	f := file.NewCSV(cfg.Filename)
	v := validator.New()
	l := logger.NewLogger()

	defer l.Sync()

	s, err := sql.NewStorage(cfg.Database, l)
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

	go app.cleanupExpiredSegments(cfg.Database.CheckInterval)

	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: app.router,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Error %s launching server", err)
	}
}

func (app *application) cleanupExpiredSegments(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		app.storage.DeleteExpiredSegments()
	}
}
