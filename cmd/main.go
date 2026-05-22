package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/RomanKovalev007/organization_service/internal/config"
	"github.com/RomanKovalev007/organization_service/internal/repository"
	"github.com/RomanKovalev007/organization_service/internal/service"
	"github.com/RomanKovalev007/organization_service/internal/transport"
	"github.com/RomanKovalev007/organization_service/internal/transport/handler"
	"github.com/RomanKovalev007/organization_service/pkg/logger"
	"github.com/RomanKovalev007/organization_service/pkg/postgres"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	l := logger.New(cfg.Log.Level)
	slog.SetDefault(l)

	db, err := postgres.New(postgres.Config{
		DSN:             cfg.DB.DSN(),
		MaxOpenConns:    cfg.DB.MaxOpenConns,
		MaxIdleConns:    cfg.DB.MaxIdleConns,
		ConnMaxLifetime: cfg.DB.ConnMaxLifetime,
	})
	if err != nil {
		slog.Error("connect to postgres", "err", err)
		os.Exit(1)
	}

	deptRepo := repository.NewDepartmentRepo(db)
	empRepo := repository.NewEmployeeRepo(db)
	txManager := repository.NewTxManager(db, deptRepo, empRepo)

	svc := service.NewService(deptRepo, empRepo, txManager)
	h := handler.New(svc)
	router := transport.NewRouter(h)

	srv := &http.Server{
		Addr:         cfg.HTTP.Addr,
		Handler:      router,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	go func() {
		slog.Info("listening", "addr", cfg.HTTP.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("listen", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.WriteTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown", "err", err)
		os.Exit(1)
	}

	slog.Info("stopped")
}
