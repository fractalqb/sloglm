package sloglm

import (
	"bytes"
	"context"
	"encoding"
	"fmt"
	"io"
	"sync"

	"git.fractalqb.de/fractalqb/sllm/v2"
	"golang.org/x/exp/slog"
)

var DefaultOptions = slog.HandlerOptions{
	Level: &slog.LevelVar{},
}

type Handler struct {
	w    io.Writer
	hdr  HeaderFunc
	opts *slog.HandlerOptions
	mu   sync.Mutex
}

func NewHandler(w io.Writer, header HeaderFunc, opts *slog.HandlerOptions) *Handler {
	if opts == nil {
		opts = &DefaultOptions
	}
	return &Handler{w: w, hdr: header, opts: opts}
}

func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	var hdr bytes.Buffer
	if h.hdr != nil {
		h.hdr(&hdr, r)
	}
	buf, err := sllm.Expand(hdr.Bytes(), r.Message, argPrinter{h, &r}.print)
	if err != nil {
		return err
	}
	buf = append(buf, '\n')
	h.mu.Lock()
	defer h.mu.Unlock()
	_, err = h.w.Write(buf)
	return err
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// TODO
	return h
}

func (h *Handler) WithGroup(name string) slog.Handler {
	// TODO
	return h
}

type argPrinter struct {
	h *Handler
	r *slog.Record
}

func (ap argPrinter) print(wr *sllm.ArgWriter, idx int, name string) (n int, err error) {
	found := false
	ap.r.Attrs(func(att slog.Attr) bool {
		if att.Key == name {
			switch att.Value.Kind() {
			case slog.KindString:
				n, err = wr.WriteString(att.Value.String())
			case slog.KindInt64:
				n, err = wr.WriteInt(att.Value.Int64())
			default:
				if tm, ok := att.Value.Any().(encoding.TextMarshaler); ok {
					var data []byte
					data, err = tm.MarshalText()
					if err == nil {
						n, err = wr.Write(data)
					}
				} else {
					n, err = fmt.Fprint(wr, att.Value.Any())
				}
			}
			found = true
			return false
		}
		return true
	})
	if !found {
		fmt.Fprintf(wr, "<arg #%d missing>", idx)
	}
	return
}
