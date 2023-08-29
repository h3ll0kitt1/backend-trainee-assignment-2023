package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/h3ll0kitt1/avitotest/internal/config"
	"github.com/h3ll0kitt1/avitotest/internal/file"
	"github.com/h3ll0kitt1/avitotest/internal/storage"
	"github.com/h3ll0kitt1/avitotest/internal/storage/sql"
	"github.com/h3ll0kitt1/avitotest/internal/validator"
)

type application struct {
	storage   storage.Storage
	router    *chi.Mux
	file      file.File
	validator validator.Validator
}

func main() {

	cfg := config.NewConfig()

	r := chi.NewRouter()
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
