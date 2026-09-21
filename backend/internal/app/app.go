package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/trade-diary/backend/internal/config"
	"github.com/trade-diary/backend/internal/domain/auth"
	"github.com/trade-diary/backend/internal/domain/column"
	"github.com/trade-diary/backend/internal/domain/journal"
	"github.com/trade-diary/backend/internal/infrastructure/postgres"
	httptransport "github.com/trade-diary/backend/internal/transport/http"
)

type App struct {
	config config.Config
	logger *slog.Logger
}

func New(cfg config.Config, logger *slog.Logger) *App {
	return &App{config: cfg, logger: logger}
}

func (a *App) Run(ctx context.Context) error {
	pool, err := pgxpool.New(ctx, a.config.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		return err
	}
	authService := auth.NewService(postgres.NewAuthRepository(pool))
	journalService := journal.NewService(postgres.NewJournalRepository(pool))
	columnService := column.NewService(postgres.NewColumnRepository(pool))

	server := &http.Server{
		Addr:              a.config.HTTPAddress,
		Handler:           httptransport.NewRouter(a.config.FrontendOrigin, authService, journalService, columnService),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		a.logger.Info("http server started", "address", a.config.HTTPAddress, "environment", a.config.Environment)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	}
}
