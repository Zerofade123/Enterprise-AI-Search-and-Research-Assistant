package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Zerofade123/Enterprise-AI-Search-and-Research-Assistant/backend/internal/auth/infra/postgres"
	"github.com/Zerofade123/Enterprise-AI-Search-and-Research-Assistant/backend/internal/auth/service"
	"github.com/Zerofade123/Enterprise-AI-Search-and-Research-Assistant/backend/internal/auth/transport/http"
	"github.com/Zerofade123/Enterprise-AI-Search-and-Research-Assistant/backend/internal/platform/config"
	"github.com/Zerofade123/Enterprise-AI-Search-and-Research-Assistant/backend/internal/platform/logging"
	"github.com/Zerofade123/Enterprise-AI-Search-and-Research-Assistant/backend/internal/platform/storage"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load("auth-service")
	if err != nil {
		panic(err)
	}

	logger, err := logging.NewLogger(cfg)
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	pool, err := storage.NewPostgresPool(cfg)
	if err != nil {
		logger.Fatal("failed to connect to postgres", zap.Error(err))
	}
	defer pool.Close()

	userRepo := postgres.NewUserRepository(pool)
	sessionRepo := postgres.NewSessionRepository(pool)
	keyRepo := postgres.NewAPIKeyRepository(pool)

	authSvc := service.NewAuthService(userRepo, sessionRepo, keyRepo, cfg, logger)
	server := http.NewServer(authSvc, logger)

	httpServer := &http.Server{
		Addr:         ":" + server.Port(),
		Handler:      server.Router(),
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	go func() {
		logger.Info("auth service started", zap.String("addr", httpServer.Addr))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("http server error", zap.Error(err))
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownGracePeriod)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", zap.Error(err))
	}

	logger.Info("auth service stopped")
}
