package sloglm

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	"golang.org/x/exp/slog"
)

func Example() {
	t := time.Date(2023, 05, 04, 20, 30, 40, 0, time.UTC)
	logger := slog.New(NewHandler(os.Stdout, nil, nil))
	logger.Info("A `level` message `at` `about` and `about.bar`",
		"about", slog.GroupValue(
			slog.Bool("foo", true),
			slog.Int("bar", 4711),
			slog.Float64("baz", 3.14159),
		),
		"level", slog.LevelInfo,
		"at", t,
		"level", 7,
	)
	// Output:
	// A `level:INFO` message `at:2023-05-04T20:30:40Z` `about:[foo=true bar=4711 baz=3.14159]` and `about.bar:4711`
}

func ExampleInt() {
	logger := slog.New(NewHandler(os.Stdout, DefaultHeader, nil))
	logger.Info("`num`", "num", 4)
	// Output:
	// _
}

func simpleHeader(w *bytes.Buffer, r slog.Record) error {
	_, err := fmt.Fprintf(w, "%s %s ",
		r.Time.Format(time.StampMilli),
		r.Level,
	)
	return err
}

func BenchmarkSimpleHeader(b *testing.B) {
	var buf bytes.Buffer
	logger := slog.New(NewHandler(&buf, simpleHeader, nil))
	for i := 0; i < b.N; i++ {
		logger.Info("added `count` x `item` to shopping cart by `user`",
			"count", 7,
			"item", "Hat",
			"user", "John Doe",
		)
	}
}

func BenchmarkDefaultHeader(b *testing.B) {
	var buf bytes.Buffer
	logger := slog.New(NewHandler(&buf, DefaultHeader, nil))
	for i := 0; i < b.N; i++ {
		logger.Info("added `count` x `item` to shopping cart by `user`",
			"count", 7,
			"item", "Hat",
			"user", "John Doe",
		)
	}
}
