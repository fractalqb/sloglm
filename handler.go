package sloglm

import (
	"bytes"
	"context"
	"encoding"
	"fmt"
	"io"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"git.fractalqb.de/fractalqb/sllm/v2"
	"golang.org/x/exp/slog"
)

var DefaultOptions = slog.HandlerOptions{
	Level: &slog.LevelVar{},
}

type Handler struct {
	w     io.Writer
	hdr   HeaderFunc
	opts  *slog.HandlerOptions
	group string
	mu    sync.Mutex
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

var printStatePool = sync.Pool{New: func() any { return new(printState) }}

func (h *Handler) Handle(ctx context.Context, r slog.Record) (err error) {
	pstat := printStatePool.Get().(*printState)
	defer printStatePool.Put(pstat)
	if h.hdr != nil {
		if pstat.buf, err = h.hdr(pstat.buf[:0], r); err != nil {
			return err
		}
	}
	pstat.init(&r)
	if h.group != "" {
		pstat.buf = append(pstat.buf, h.group...)
		pstat.buf = append(pstat.buf, ". "...)
	}
	if h.opts.AddSource && r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		pstat.buf = append(pstat.buf, '@')
		pstat.buf = append(pstat.buf, f.File...)
		pstat.buf = append(pstat.buf, ':')
		pstat.buf = strconv.AppendInt(pstat.buf, int64(f.Line), 10)
		pstat.buf = append(pstat.buf, ' ')
	}
	if r.NumAttrs() == 0 {
		pstat.buf = append(pstat.buf, r.Message...)
	} else {
		pstat.buf, err = sllm.Expand(pstat.buf, r.Message, pstat.print)
		if err != nil {
			return err
		}
		if pstat.formNo < len(pstat.atts) {
			pstat.buf = append(pstat.buf, " ("...)
			w := bytes.NewBuffer(pstat.buf)
			sep := ""
			for i := 0; i < len(pstat.atts); i++ {
				if pstat.formArgs&(1<<i) == 0 {
					w.WriteString(sep)
					fmt.Fprint(w, pstat.atts[i])
					sep = " "
				}
			}
			pstat.buf = append(w.Bytes(), ')')
		}
	}
	pstat.buf = append(pstat.buf, '\n')
	h.mu.Lock()
	defer h.mu.Unlock()
	_, err = h.w.Write(pstat.buf)
	return err
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// TODO Handler.WithAttrs()
	return h
}

func (h *Handler) WithGroup(name string) slog.Handler {
	return &Handler{
		w:     h.w,
		hdr:   h.hdr,
		opts:  h.opts,
		group: h.group + "." + name,
	}
}

type printState struct {
	buf      []byte
	atts     []slog.Attr
	formArgs uint64 // TODO what if there are more args
	formNo   int
}

func (ps *printState) init(r *slog.Record) {
	ps.atts = ps.atts[:0]
	r.Attrs(func(a slog.Attr) bool {
		ps.atts = append(ps.atts, a)
		return true
	})
	ps.formArgs = 0
	ps.formNo = 0
}

func (ps printState) attName(key string) (int, slog.Attr) {
	for i, att := range ps.atts {
		if att.Key == key {
			return i, att
		}
	}
	return -1, slog.Attr{}
}

func (ps printState) attPath(idx int, path string) (int, slog.Attr) {
	if path == "." {
		return -1, slog.Attr{}
	}
	names := strings.Split(path[1:], ".")
	var att slog.Attr
	if idx < len(ps.atts) && ps.atts[idx].Key == names[0] {
		att = ps.atts[idx]
	} else {
		var i int
		if i, att = ps.attName(names[0]); i < 0 {
			return -1, slog.Attr{}
		}
		idx = i
	}
NEXT_NAME:
	for _, name := range names[1:] {
		if att.Value.Kind() != slog.KindGroup {
			return -1, slog.Attr{}
		}
		for _, a := range att.Value.Group() {
			if a.Key == name {
				att = a
				continue NEXT_NAME
			}
		}
		return -1, slog.Attr{}
	}
	return idx, att
}

func (ap *printState) print(wr *sllm.ArgWriter, idx int, name string) (n int, err error) {
	var att slog.Attr
	if name[0] == '.' { // TODO sllm does not support empty names – riksy?
		var i int
		if i, att = ap.attPath(idx, name); i < 0 {
			return 0, fmt.Errorf("no argument %d", idx)
		}
		idx = i
	} else if idx < len(ap.atts) && ap.atts[idx].Key == name {
		att = ap.atts[idx]
	} else {
		var i int
		if i, att = ap.attName(name); i < 0 {
			return 0, fmt.Errorf("no argument %d", idx)
		}
		idx = i
	}
	flag := uint64(1) << uint64(idx)
	if ap.formArgs&flag == 0 {
		ap.formArgs |= flag
		ap.formNo++
	}
	switch att.Value.Kind() {
	case slog.KindString:
		n, err = wr.WriteString(att.Value.String())
	case slog.KindInt64:
		n, err = wr.WriteInt64(att.Value.Int64())
	// case slog.KindTime:
	// 	n, err = wr.WriteTime(att.Value.Time(), sllm.Tyear|sllm.Tmillis)
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
	return
}
