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
		want  string
		input time.Duration
	}{
		{want: "450ms", input: 450 * time.Millisecond},
		{want: "0ms", input: 0},
		{want: "999ms", input: 999 * time.Millisecond},
		{want: "1.0s", input: 1 * time.Second},
		{want: "3.2s", input: 3200 * time.Millisecond},
		{want: "59.0s", input: 59 * time.Second},
		{want: "1m00s", input: 1 * time.Minute},
		{want: "2m03s", input: 2*time.Minute + 3*time.Second},
		{want: "500ms", input: -500 * time.Millisecond},
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
		want string
		n    int64
	}{
		{name: "below threshold", want: "42", n: 42},
		{name: "negative below threshold", want: "-100", n: -100},
		{name: "at threshold boundary", want: "9999", n: 9999},
		{name: "above threshold", want: "10,000", n: 10_000},
		{name: "large number", want: "1,234,567", n: 1_234_567},
		{name: "negative above threshold", want: "-50,000", n: -50_000},
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
		want string
		pct  Pct
	}{
		{name: "whole number", want: "42%", pct: Pct(42)},
		{name: "decimal", want: "94.3%", pct: Pct(94.3)},
		{name: "zero", want: "0%", pct: Pct(0)},
		{name: "hundred", want: "100%", pct: Pct(100)},
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
