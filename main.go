package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"

	"github.com/Xevion/motophoto/internal/server"
)

// shouldUseJSON returns true when structured JSON logging should be used.
//
// Priority:
//  1. LOG_JSON env var explicitly set → use that value
//  2. Otherwise → pretty (the Docker entrypoint sets LOG_JSON=true explicitly)
func shouldUseJSON() bool {
	v, ok := os.LookupEnv("LOG_JSON")
	if ok {
		return v == "true" || v == "1"
	}
	return false
}

// parseLogLevel converts a LOG_LEVEL env value to an slog.Level.
// Falls back to Info if unset or unrecognized.
func parseLogLevel() slog.Level {
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func initLogging() {
	level := parseLogLevel()

	var handler slog.Handler
	if shouldUseJSON() {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	} else {
		handler = tint.NewHandler(os.Stderr, &tint.Options{
			Level:      level,
			TimeFormat: time.TimeOnly,
		})
	}

	slog.SetDefault(slog.New(handler))
}

func main() {
	// Load .env if present (ignored in production where env vars are set directly)
	_ = godotenv.Load()

	initLogging()

	srv, err := server.New()
	if err != nil {
		slog.Error("failed to create server", "error", err)
		os.Exit(1)
	}

	httpServer := &http.Server{
		Addr:         srv.Addr(),
		Handler:      srv.Router(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("server starting", "addr", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-done
	slog.Info("server shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	srv.Close()
	slog.Info("server stopped")
}
