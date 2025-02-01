package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"wallet/internal/config"
	"wallet/internal/http-server/handlers/getter"
	"wallet/internal/http-server/handlers/transaction"
	mwLogger "wallet/internal/http-server/middleware/logger"
	"wallet/internal/lib/logger/sl"
	"wallet/storage/postgresql"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// fmt.Println("Successfully connected to PostgreSQL")
func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting wallet", slog.String("env", cfg.Env))

	var dbURL string = fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.SSLMode,
	)

	storage, err := postgresql.NewStorage(dbURL)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	log.Info("Successfully connected to PostgreSQL")

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(mwLogger.New(log))
	router.Use(middleware.URLFormat)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)


	router.Get("/api/v1/wallets/{WALLET_UUID}", getter.FetchWallet(log, storage))
	router.Post("/api/v1/wallet", transaction.WalletOperation(log, storage))

	log.Info("starting server", slog.String("address", cfg.Address))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}

	log.Error("server stopped")

}

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
