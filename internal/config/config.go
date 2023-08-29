package config

import (
	"flag"
	"os"
)

type Config struct {
	Addr     string
	Database string
	Filename string
}

func NewConfig() *Config {

	var (
		flagRunAddr      string
		flagDatabasePath string
		flagFilename     string
	)

	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&flagDatabasePath, "d", "", "sql database to store metrics")
	flag.StringVar(&flagFilename, "f", "/tmp/file.csv", "file to download history from app")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		flagRunAddr = envRunAddr
	}

	if envDatabasePath := os.Getenv("DATABASE_DSN"); envDatabasePath != "" {
		flagDatabasePath = envDatabasePath
	}

	if envFilename := os.Getenv("FILENAME"); envFilename != "" {
		flagFilename = envFilename
	}

	addr := flagRunAddr
	database := flagDatabasePath
	filename := flagFilename

	return &Config{
		Addr:     addr,
		Database: database,
		Filename: filename,
	}
}
