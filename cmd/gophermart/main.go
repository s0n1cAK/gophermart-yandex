package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	config "yandex-diplom/internal/config/gophermart"
	"yandex-diplom/internal/gophermart"
	"yandex-diplom/internal/job"
	"yandex-diplom/internal/logger"
	"yandex-diplom/internal/server"
	"yandex-diplom/internal/storage/postgresql"
	"yandex-diplom/internal/worker"

	"go.uber.org/zap"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger, err := logger.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to init logger:%s", err)
		os.Exit(1)
	}
	defer func() {
		_ = logger.Sync()
	}()

	cfg, err := config.New()
	if err != nil {
		logger.Fatal("Failed to parse config:", zap.Error(err))
	}

	storage, err := postgresql.NewPostgresStorage(cfg.DatabaseURI)
	if err != nil {
		logger.Fatal("Failed to connect to database:", zap.Error(err))
	}
	defer storage.Database.Close()

	service := gophermart.New(storage, logger, cfg.Environment, cfg.AccuralAddress)

	jobCh := make(chan job.Job, 100)

	workers := worker.InitWorkers(ctx, 4, service, jobCh)
	workers.StartOrderProcessor(worker.OrderConfig{BatchSize: 30})
	workers.StartBalanceProcessor(worker.BalanceConfig{})

	srv, err := server.New(cfg, service)
	if err != nil {
		logger.Fatal("failed to create server", zap.Error(err))
	}

	go func() {
		srv.Start()
	}()

	<-ctx.Done()
	logger.Info("shutting down...")

	shutdownCtx, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Warn("http shutdown error", zap.Error(err))
	}
}
