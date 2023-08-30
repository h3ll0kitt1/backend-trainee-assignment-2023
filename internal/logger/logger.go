package logger

import (
	"go.uber.org/zap"
)

func NewLogger() *zap.SugaredLogger {
	atom := zap.NewAtomicLevel()
	atom.SetLevel(zap.InfoLevel)

	cfg := zap.NewProductionConfig()
	cfg.Level = atom
	logger := zap.Must(cfg.Build())
	sugaredLogger := logger.Sugar()
	return sugaredLogger
}
