package logging

import (
	"strings"

	"github.com/Zerofade123/Enterprise-AI-Search-and-Research-Assistant/backend/internal/platform/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(cfg *config.AppConfig) (*zap.Logger, error) {
	level := zap.NewAtomicLevel()
	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		level.SetLevel(zap.DebugLevel)
	case "info":
		level.SetLevel(zap.InfoLevel)
	case "warn":
		level.SetLevel(zap.WarnLevel)
	case "error":
		level.SetLevel(zap.ErrorLevel)
	default:
		level.SetLevel(zap.InfoLevel)
	}

	zapCfg := zap.NewProductionConfig()
	zapCfg.Level = level
	zapCfg.Encoding = "json"
	zapCfg.EncoderConfig.TimeKey = "timestamp"
	zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := zapCfg.Build()
	if err != nil {
		return nil, err
	}

	return logger.With(
		zap.String("service", cfg.ServiceName),
		zap.String("env", cfg.Env),
	), nil
}
