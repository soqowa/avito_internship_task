package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/user/reviewer-svc/internal/infrastructure/db/postgres"
	"github.com/user/reviewer-svc/internal/app"
	"github.com/user/reviewer-svc/internal/app/config"
	"github.com/user/reviewer-svc/internal/infrastructure/logger"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	lg := logger.New(cfg.LogLevel)

	if err := postgres.Migrate(cfg.DBDSN, "migrations"); err != nil {
		lg.Error("migrations failed", "err", err)
		os.Exit(1)
	}

	dbCtx, cancelDB := context.WithTimeout(ctx, 10*time.Second)
	defer cancelDB()

	pool, err := postgres.NewPool(dbCtx, cfg.DBDSN)
	if err != nil {
		lg.Error("db connect failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	handler := app.NewHandler(r, pool, lg)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: handler,
	}

	go func() {
		lg.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			lg.Error("server failed", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	shCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shCtx); err != nil {
		lg.Error("server shutdown error", "err", err)
	}
}
