package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/trade-diary/backend/internal/config"
	"github.com/trade-diary/backend/internal/domain/analytics"
	"github.com/trade-diary/backend/internal/domain/attachment"
	"github.com/trade-diary/backend/internal/domain/auth"
	"github.com/trade-diary/backend/internal/domain/column"
	"github.com/trade-diary/backend/internal/domain/journal"
	"github.com/trade-diary/backend/internal/domain/observation"
	"github.com/trade-diary/backend/internal/infrastructure/postgres"
	localstorage "github.com/trade-diary/backend/internal/infrastructure/storage"
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
	started := time.Now()
	a.logger.InfoContext(ctx, "Запуск приложения", "event", "app.starting", "address", a.config.HTTPAddress)
	pool, err := pgxpool.New(ctx, a.config.DatabaseURL)
	if err != nil {
		return errors.New("некорректные настройки PostgreSQL: проверьте DATABASE_URL")
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("проверка соединения с PostgreSQL: %w", err)
	}
	a.logger.InfoContext(ctx, "Соединение с PostgreSQL проверено", "event", "database.connected", "duration_ms", float64(time.Since(started).Microseconds())/1000)
	authService := auth.NewService(postgres.NewAuthRepository(pool))
	journalService := journal.NewService(postgres.NewJournalRepository(pool))
	columnService := column.NewService(postgres.NewColumnRepository(pool))
	observationService := observation.NewService(postgres.NewObservationRepository(pool))
	storage, err := localstorage.NewLocal(a.config.UploadDirectory)
	if err != nil {
		return fmt.Errorf("подготовка каталога скриншотов: %w", err)
	}
	a.logger.InfoContext(ctx, "Хранилище скриншотов готово", "event", "storage.ready", "directory", a.config.UploadDirectory)
	attachmentService := attachment.NewService(postgres.NewAttachmentRepository(pool), storage)
	analyticsService := analytics.NewService(postgres.NewAnalyticsRepository(pool))

	server := &http.Server{
		Addr:              a.config.HTTPAddress,
		Handler:           httptransport.NewRouter(a.logger, a.config.FrontendOrigin, a.config.Environment, authService, journalService, columnService, observationService, attachmentService, analyticsService),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	listener, err := net.Listen("tcp", a.config.HTTPAddress)
	if err != nil {
		return fmt.Errorf("открытие HTTP-порта %s: %w", a.config.HTTPAddress, err)
	}
	defer listener.Close()
	a.logger.InfoContext(ctx, "HTTP-сервер готов принимать запросы", "event", "http.server.started", "address", listener.Addr().String(), "startup_duration_ms", float64(time.Since(started).Microseconds())/1000)
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.Serve(listener)
	}()

	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("работа HTTP-сервера: %w", err)
	case <-ctx.Done():
		a.logger.Info("Получен сигнал остановки, завершаем активные HTTP-запросы", "event", "app.stopping", "shutdown_timeout_ms", 10000)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("завершение активных HTTP-запросов: %w", err)
		}
		a.logger.Info("HTTP-сервер остановлен", "event", "http.server.stopped")
		return nil
	}
}
