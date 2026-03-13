package middleware

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestLogLevel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		path string
		want slog.Level
	}{
		{"/api/v1/events", slog.LevelInfo},
		{"/api/health", slog.LevelInfo},
		{"/src/main.ts", LevelTrace},
		{"/node_modules/svelte/index.js", LevelTrace},
		{"/@vite/client", LevelTrace},
		{"/.svelte-kit/generated/root.svelte", LevelTrace},
		{"/styled-system/tokens/index.mjs", LevelTrace},
		{"/about", slog.LevelDebug},
		{"/", slog.LevelDebug},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()
			got := requestLogLevel(tt.path)
			assert.Equal(t, tt.want, got)
		})
	}
}
