package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/mattsre/badger/internal/circleci"
	"github.com/mattsre/badger/internal/handler"
)

func main() {
	addr := envOrDefault("BADGER_ADDR", "127.0.0.1:8080")
	token := os.Getenv("CIRCLECI_TOKEN")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{
		Addr:    addr,
		Handler: handler.New(circleci.NewClient(token)),
	}

	go func() {
		slog.Info("badger listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down", "cause", context.Cause(ctx))

	if err := srv.Shutdown(context.Background()); err != nil {
		slog.Error("shutdown failed", "err", err)
		os.Exit(1)
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
