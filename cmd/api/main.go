package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/sakid00/enmasse-be/internal/config"
	apphttp "github.com/sakid00/enmasse-be/internal/http"
	"github.com/sakid00/enmasse-be/internal/mail"
	"github.com/sakid00/enmasse-be/internal/service"
	"github.com/sakid00/enmasse-be/internal/store"
	"github.com/sakid00/enmasse-be/internal/turnstile"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	level := slog.LevelInfo
	if !cfg.IsProduction() {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})))

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	db, err := store.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("db connect", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		slog.Error("goose dialect", "err", err)
		os.Exit(1)
	}
	sqlDB := stdlib.OpenDBFromPool(db.Pool)
	if err := goose.Up(sqlDB, "migrations"); err != nil {
		sqlDB.Close()
		slog.Error("migrations", "err", err)
		os.Exit(1)
	}
	sqlDB.Close()

	authSvc := service.NewAuthService(db, cfg, turnstile.New(cfg), mail.New(cfg))
	handler := apphttp.NewRouter(authSvc, apphttp.RouterOptions{
		CORSOrigins:      cfg.CORSOrigins,
		JWTAccessSecret:  cfg.JWTAccessSecret,
		JWTIssuer:        cfg.JWTIssuer,
		JWTServiceSecret: cfg.JWTServiceSecret,
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		slog.Info("server starting", "port", cfg.Port, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down")
	shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutCancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		slog.Error("shutdown error", "err", err)
	}
}
