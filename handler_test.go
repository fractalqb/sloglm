package sloglm

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"git.fractalqb.de/fractalqb/sllm/v3"
)

func Example() {
	t := time.Date(2023, 05, 04, 20, 30, 40, 0, time.UTC)
	logger := slog.New(NewHandler(os.Stdout, nil, nil))
	logger.Info("A `level` message `at` `about` and `.about.bar`",
		"level", slog.LevelInfo,
		"at", t,
		"level", 7,
		"about", slog.GroupValue(
			slog.Bool("foo", true),
			slog.Int("bar", 4711),
			slog.Float64("baz", 3.14159),
		),
	)
	// Output:
	// A `level:INFO` message `at:2023-05-04T20:30:40Z` `about:[foo=true bar=4711 baz=3.14159]` and `.about.bar:4711` (level=7)
}

func simpleHeader(w []byte, r slog.Record) ([]byte, error) {
	buf := bytes.NewBuffer(w)
	_, err := fmt.Fprintf(buf, "%s %s ",
		r.Time.Format(time.StampMilli),
		r.Level,
	)
	return buf.Bytes(), err
}

func BenchmarkSimpleHeader(b *testing.B) {
	var buf bytes.Buffer
	logger := slog.New(NewHandler(&buf, simpleHeader, nil))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		logger.Info("added `count` x `item` to shopping cart by `user`",
			"count", 7,
			"item", "Hat",
			"user", "John Doe",
		)
	}
}

func BenchmarkDefaultHeader(b *testing.B) {
	var buf bytes.Buffer
	logger := slog.New(NewHandler(&buf, DefaultHeader(), nil))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		logger.Info("added `count` x `item` to shopping cart by `user`",
			"count", 7,
			"item", "Hat",
			"user", "John Doe",
		)
	}
}

func BenchmarkDefaultHeaderFast(b *testing.B) {
	var buf bytes.Buffer
	logger := slog.New(NewHandler(&buf, DefaultHeader(), nil))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		logger.LogAttrs(context.Background(), slog.LevelInfo,
			"added `count` x `item` to shopping cart by `user`",
			slog.Int("count", 7),
			slog.String("item", "Hat"),
			slog.String("user", "John Doe"),
		)
	}
}

func BenchmarkSlogText(b *testing.B) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		logger.Info("user added items to shopping cart",
			"count", 7,
			"item", "Hat",
			"user", "John Doe",
		)
	}
}

func BenchmarkSlogJSON(b *testing.B) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		logger.Info("user added items to shopping cart",
			"count", 7,
			"item", "Hat",
			"user", "John Doe",
		)
	}
}

func Example_compareOutputs() {
	const form = "added `count` x `item` to shopping cart by `user`"
	args := []any{"count", 7, "item", "Hat", "user", "John Doe"}
	opts := slog.HandlerOptions{Level: &slog.LevelVar{}, AddSource: true}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &opts))
	fmt.Println("Std Text:")
	logger.Info(form, args...)
	logger.WithGroup("gruppe").Info(form, args...)
	logger = slog.New(slog.NewJSONHandler(os.Stdout, &opts))
	fmt.Println("Std JSON:")
	logger.Info(form, args...)
	logger.WithGroup("gruppe").Info(form, args...)

	logger = slog.New(NewHandler(
		os.Stdout,
		NewHeader(sllm.Tdefault|sllm.Tyear).Append,
		&opts,
	))
	fmt.Println("sllm:")
	logger.Info(form, args...)
	logger.WithGroup("gruppe").Info(form, args...)
	// DISABLED Output:
	// time=2023-05-06T13:24:00.476+02:00 level=INFO msg="added `count` x `item` to shopping cart by `user`" count=7 item=Hat user="John Doe"
	// {"time":"2023-05-06T13:24:00.476762195+02:00","level":"INFO","msg":"added `count` x `item` to shopping cart by `user`","count":7,"item":"Hat","user":"John Doe"}
	// 2023-05-06 Sa 13:24:00 INFO	added `count:7` x `item:Hat` to shopping cart by `user:John Doe`
}

func BenchmarkAttrLookup(b *testing.B) {
	const form = "A `level` message `at` `about` with `duration` some `p1`, `p2`, `p3`, `p4`, `p5`"
	var (
		level    = slog.Int("level", 7)
		at       = slog.Time("at", time.Now())
		about    = slog.String("about", "something")
		duration = slog.Duration("duration", 5*time.Minute)
		p1       = slog.Int("p1", 1)
		p2       = slog.Int("p2", 1)
		p3       = slog.Int("p3", 1)
		p4       = slog.Int("p4", 1)
		p5       = slog.Int("p5", 1)
	)
	var buf bytes.Buffer
	logger := slog.New(NewHandler(&buf, DefaultHeader(), nil))
	b.ResetTimer()

	b.Run("index match", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf.Reset()
			logger.LogAttrs(context.Background(), slog.LevelInfo, form,
				level, at, about, duration,
				p1, p2, p3, p4, p5,
			)
		}
	})

	b.Run("index mismatch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf.Reset()
			logger.LogAttrs(context.Background(), slog.LevelInfo, form,
				p1, p5, p4, p3, p2,
				duration, about, at, level,
			)
		}
	})
}
