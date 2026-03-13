package logging

import (
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatDuration(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input time.Duration
		want  string
	}{
		{450 * time.Millisecond, "450ms"},
		{0, "0ms"},
		{999 * time.Millisecond, "999ms"},
		{1 * time.Second, "1.0s"},
		{3200 * time.Millisecond, "3.2s"},
		{59 * time.Second, "59.0s"},
		{1 * time.Minute, "1m00s"},
		{2*time.Minute + 3*time.Second, "2m03s"},
		{-500 * time.Millisecond, "500ms"},
	}
	for _, tt := range tests {
		t.Run(tt.input.String(), func(t *testing.T) {
			t.Parallel()
			got := formatDuration(slog.DurationValue(tt.input))
			assert.Equal(t, tt.want, got.String())
		})
	}
}

func TestFormatInt(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		n    int64
		want string
	}{
		{"below threshold", 42, "42"},
		{"negative below threshold", -100, "-100"},
		{"at threshold boundary", 9999, "9999"},
		{"above threshold", 10_000, "10,000"},
		{"large number", 1_234_567, "1,234,567"},
		{"negative above threshold", -50_000, "-50,000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := formatInt(slog.Int64Value(tt.n))
			if tt.n > -commaThreshold && tt.n < commaThreshold {
				// Should pass through unchanged as int64
				assert.Equal(t, slog.KindInt64, got.Kind())
			} else {
				assert.Equal(t, tt.want, got.String())
			}
		})
	}
}

func TestPct_LogValue(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		pct  Pct
		want string
	}{
		{"whole number", Pct(42), "42%"},
		{"decimal", Pct(94.3), "94.3%"},
		{"zero", Pct(0), "0%"},
		{"hundred", Pct(100), "100%"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.pct.LogValue()
			assert.Equal(t, tt.want, got.String())
		})
	}
}

func TestFormatters_ReturnsThreeFormatters(t *testing.T) {
	t.Parallel()
	fmts := Formatters()
	assert.Len(t, fmts, 3)
}
