package sloglm

import (
	"strings"
	"sync"
	"sync/atomic"

	"git.fractalqb.de/fractalqb/sllm/v2"
	"golang.org/x/exp/slog"
)

type HeaderFunc func([]byte, slog.Record) ([]byte, error)

type Header struct {
	time     sllm.TimeFormat
	lvlLen   int32
	lvlFill  atomic.Value
	lvlMutex sync.Mutex
}

func NewHeader(t sllm.TimeFormat) *Header {
	h := &Header{time: t, lvlLen: 0}
	h.lvlFill.Store("")
	return h
}

func (hdr *Header) Append(w []byte, r slog.Record) ([]byte, error) {
	(*sllm.ArgWriter)(&w).WriteTime(r.Time, hdr.time)
	w = append(w, " ["...)
	level := r.Level.String()
	w = append(w, level...)
	w = append(w, "] "...)
	if ll := atomic.LoadInt32(&hdr.lvlLen); len(level) < int(ll) {
		fill := hdr.lvlFill.Load().(string)
		w = append(w, fill[:int(ll)-len(level)]...)
	} else if len(level) > int(ll) {
		hdr.lvlMutex.Lock()
		fill := hdr.lvlFill.Load().(string)
		if l := len(level); l > len(fill) {
			fill = strings.Repeat(" ", l)
			hdr.lvlFill.Store(fill)
			atomic.StoreInt32(&hdr.lvlLen, int32(l))
		}
		hdr.lvlMutex.Unlock()
	}
	return w, nil
}

func DefaultHeader() HeaderFunc { return NewHeader(sllm.Tdefault).Append }
