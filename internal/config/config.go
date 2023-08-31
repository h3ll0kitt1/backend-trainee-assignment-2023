package config

import (
	"errors"
	"flag"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Addr     string
	Filename string
	Database Database
}

type Database struct {
	POSTGRES_DB       string
	POSTGRES_USER     string
	POSTGRES_PASSWORD string
	POSTGRES_PORT     string
	DATABASE_HOST     string
	CheckInterval     time.Duration
}

func NewConfig() (*Config, error) {

	var (
		flagCheckInterval int
		flagRunAddr       string
		flagFilename      string
		flagDatabaseHost  string
	)

	var (
		envPOSTGRES_DB       string
		envPOSTGRES_USER     string
		envPOSTGRES_PORT     string
		envPOSTGRES_PASSWORD string
	)

	flag.IntVar(&flagCheckInterval, "r", 1, "number of minuts to sync expired segments")
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&flagDatabaseHost, "d", "localhost", "host to run database")
	flag.StringVar(&flagFilename, "f", "/tmp/file.csv", "file to download history from app")
	flag.Parse()

	envCheckInterval, err := strconv.Atoi(os.Getenv("CHECK_INTERVAL"))
	if err == nil {
		flagCheckInterval = envCheckInterval
	}

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		flagRunAddr = envRunAddr
	}

	if envDatabaseHost := os.Getenv("DATABASE_HOST"); envDatabaseHost != "" {
		flagDatabaseHost = envDatabaseHost
	}

	if envPOSTGRES_DB = os.Getenv("POSTGRES_DB"); envPOSTGRES_DB == "" {
		return nil, errors.New("Could not find ENV variable POSTGRES_DB")
	}

	if envPOSTGRES_USER = os.Getenv("POSTGRES_USER"); envPOSTGRES_USER == "" {
		return nil, errors.New("Could not find ENV variable POSTGRES_USER")
	}

	if envPOSTGRES_PORT = os.Getenv("POSTGRES_PORT"); envPOSTGRES_PORT == "" {
		return nil, errors.New("Could not find ENV variable POSTGRES_PORT")
	}

	if envPOSTGRES_PASSWORD = os.Getenv("POSTGRES_PASSWORD"); envPOSTGRES_PASSWORD == "" {
		return nil, errors.New("Could not find ENV variable POSTGRES_PASSWORD")
	}

	if envFilename := os.Getenv("FILENAME"); envFilename != "" {
		flagFilename = envFilename
	}

	addr := flagRunAddr
	filename := flagFilename
	checkInterval := time.Duration(flagCheckInterval) * time.Minute

	database := Database{
		POSTGRES_DB:       envPOSTGRES_DB,
		POSTGRES_USER:     envPOSTGRES_USER,
		POSTGRES_PORT:     envPOSTGRES_PORT,
		POSTGRES_PASSWORD: envPOSTGRES_PASSWORD,
		DATABASE_HOST:     flagDatabaseHost,
		CheckInterval:     checkInterval,
	}

	return &Config{
		Addr:     addr,
		Database: database,
		Filename: filename,
	}, nil
}
