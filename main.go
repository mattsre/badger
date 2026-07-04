package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/mattsre/badger/internal/circleci"
	"github.com/mattsre/badger/internal/handler"
)

const (
	allowedProjectsEnv = "BADGER_ALLOWED_PROJECTS"
	allowedBranchesEnv = "BADGER_ALLOWED_BRANCHES"
)

func main() {
	port := envOrDefault("PORT", "8080")
	token := envOrDefault("CIRCLECI_TOKEN", "")
	allowedProjects := csvEnv(allowedProjectsEnv)
	allowedBranches := csvEnvOrDefault(allowedBranchesEnv, []string{"main"})
	if len(allowedProjects) == 0 {
		slog.Warn("no CircleCI projects are allowed", "env", allowedProjectsEnv)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	addr := fmt.Sprintf(":%s", port)
	srv := &http.Server{
		Addr:    addr,
		Handler: handler.New(circleci.NewClient(token), allowedProjects, allowedBranches),
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

func csvEnv(key string) []string {
	return parseCSV(os.Getenv(key))
}

func csvEnvOrDefault(key string, fallback []string) []string {
	values := csvEnv(key)
	if len(values) == 0 {
		return fallback
	}
	return values
}

func parseCSV(raw string) []string {
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		values = append(values, value)
	}
	return values
}
