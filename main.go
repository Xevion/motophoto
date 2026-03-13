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
	slogformatter "github.com/samber/slog-formatter"

	"github.com/Xevion/motophoto/internal/database"
	"github.com/Xevion/motophoto/internal/logging"
	"github.com/Xevion/motophoto/internal/middleware"
	"github.com/Xevion/motophoto/internal/server"
	"github.com/Xevion/motophoto/internal/session"
	"github.com/Xevion/motophoto/internal/shutdown"
)

// shouldUseJSON returns true when structured JSON logging should be used.
//
// Priority:
//  1. LOG_JSON env var explicitly set -> use that value
//  2. Otherwise -> pretty (the Docker entrypoint sets LOG_JSON=true explicitly)
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
	case "trace":
		return middleware.LevelTrace
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
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.LevelKey && len(groups) == 0 {
					if lvl, ok := a.Value.Any().(slog.Level); ok && lvl <= middleware.LevelTrace {
						return tint.Attr(13, slog.String(a.Key, "TRC")) // 13 = bright magenta
					}
				}
				return a
			},
		})
	}

	formatted := slogformatter.NewFormatterHandler(logging.Formatters()...)(handler)

	filtered := &filteringHandler{
		inner: formatted,
		suppress: []string{
			"Unsolicited response received on idle HTTP channel",
		},
	}

	slog.SetDefault(slog.New(filtered))

	// Redirect stdlib log.Print calls (used by net/http internals and
	// third-party libraries like scs) through slog at WARN level.
	slog.SetLogLoggerLevel(slog.LevelWarn)
}

// filteringHandler suppresses known-noisy log messages from stdlib internals.
type filteringHandler struct {
	inner    slog.Handler
	suppress []string
}

func (h *filteringHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *filteringHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, s := range h.suppress {
		if strings.Contains(r.Message, s) {
			return nil
		}
	}
	return h.inner.Handle(ctx, r)
}

func (h *filteringHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &filteringHandler{inner: h.inner.WithAttrs(attrs), suppress: h.suppress}
}

func (h *filteringHandler) WithGroup(name string) slog.Handler {
	return &filteringHandler{inner: h.inner.WithGroup(name), suppress: h.suppress}
}

func main() {
	// Load .env if present (ignored in production where env vars are set directly)
	_ = godotenv.Load()

	initLogging()

	production := os.Getenv("ENVIRONMENT") == "production"
	slog.Info("environment", "production", production)

	ctx := context.Background()

	slog.Info("connecting to database", "host", database.Host())
	pool, err := database.New(ctx)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	slog.Info("database connected", "host", database.Host())

	slog.Info("running migrations")
	if err = database.Migrate(ctx, pool); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	sessions := session.New(pool, production)
	tracker := shutdown.NewTracker()

	srv, err := server.New(pool, sessions, tracker, server.Options{})
	if err != nil {
		slog.Error("failed to create server", "error", err)
		os.Exit(1)
	}

	httpServer := &http.Server{
		Addr:              srv.Addr(),
		Handler:           srv.Router(),
		ErrorLog:          slog.NewLogLogger(slog.Default().Handler(), slog.LevelWarn),
		ReadHeaderTimeout: 15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("server starting", "addr", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			done <- syscall.SIGTERM
		}
	}()

	<-done
	slog.Info("server shutting down")

	// Phase 1: stop accepting new connections and signal components.
	tracker.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		slog.Error("http server shutdown error", "error", err)
	}

	// Phase 2: wait for in-flight critical operations.
	if ok := tracker.Wait(30 * time.Second); !ok {
		slog.Warn("shutdown timed out waiting for in-flight operations")
	}

	srv.Close()
	slog.Info("server stopped")
}
