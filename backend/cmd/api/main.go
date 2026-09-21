package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/trade-diary/backend/internal/app"
	"github.com/trade-diary/backend/internal/config"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config.FromEnv()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	application := app.New(cfg, logger)
	if err := application.Run(ctx); err != nil {
		logger.Error("application stopped", "error", err)
		os.Exit(1)
	}
}
