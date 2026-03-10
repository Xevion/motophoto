// Package logging provides shared slog formatters for human-readable output.
//
// The Formatters function returns middleware formatters that transform:
//   - Durations into tiered human strings (450ms, 3.2s, 2m03s)
//   - Large integers into comma-separated strings (1,234,567)
//   - Pct values into percentage strings (42.1%)
//
// Small integers (below the comma threshold) pass through unchanged, keeping
// batch numbers, worker IDs, port numbers, etc. unformatted.
package logging

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/dustin/go-humanize"
	slogformatter "github.com/samber/slog-formatter"
)

// commaThreshold is the minimum absolute value at which integers get
// comma formatting. Values below this pass through as plain numbers.
const commaThreshold = 10_000

// Pct wraps a float64 for percentage display in structured logs.
// Pass it via slog.Any: slog.Any("match_rate", logging.Pct(94.3))
type Pct float64

func (p Pct) LogValue() slog.Value {
	v := float64(p)
	if v == float64(int64(v)) {
		return slog.StringValue(fmt.Sprintf("%d%%", int64(v)))
	}
	return slog.StringValue(fmt.Sprintf("%.1f%%", v))
}

// Formatters returns the slog-formatter middlewares for duration, integer,
// and percentage formatting.
func Formatters() []slogformatter.Formatter {
	return []slogformatter.Formatter{
		slogformatter.FormatByKind(slog.KindDuration, formatDuration),
		slogformatter.FormatByKind(slog.KindInt64, formatInt),
		slogformatter.FormatByType(func(p Pct) slog.Value {
			return p.LogValue()
		}),
	}
}

func formatDuration(v slog.Value) slog.Value {
	d := v.Duration()
	if d < 0 {
		d = -d
	}
	var s string
	switch {
	case d < time.Second:
		s = fmt.Sprintf("%dms", d.Milliseconds())
	case d < time.Minute:
		s = fmt.Sprintf("%.1fs", d.Seconds())
	default:
		m := int(d.Minutes())
		sec := int(d.Seconds()) % 60
		s = fmt.Sprintf("%dm%02ds", m, sec)
	}
	return slog.StringValue(s)
}

func formatInt(v slog.Value) slog.Value {
	n := v.Int64()
	if n > -commaThreshold && n < commaThreshold {
		return v
	}
	return slog.StringValue(humanize.Comma(n))
}
