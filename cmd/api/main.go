package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hareeshkakita/gopay/internal/config"
	"github.com/hareeshkakita/gopay/internal/db"
	"github.com/hareeshkakita/gopay/internal/repository"
	"github.com/hareeshkakita/gopay/internal/service"
	httptransport "github.com/hareeshkakita/gopay/internal/transport/http"
)

func main() {
	cfg := config.Load()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	pool, err := db.NewPool(cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to open db", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.PingContext(context.Background()); err != nil {
		logger.Error("failed to ping db", "error", err)
		os.Exit(1)
	}

	logger.Info("database connected", "env", cfg.Environment)

	walletRepo := repository.NewWalletRepository(pool)
	transactionRepo := repository.NewTransactionRepository(pool)
	ledgerRepo := repository.NewLedgerRepository(pool)
	svc := service.NewWalletService(pool, walletRepo, transactionRepo, ledgerRepo)

	router := httptransport.NewRouter(httptransport.RouterConfig{
		Logger:  logger,
		Handler: httptransport.NewHandler(svc),
	})

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.ServerPort),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		logger.Info("starting server", "addr", srv.Addr, "env", cfg.Environment)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	logger.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped")

}
