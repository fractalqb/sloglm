package sloglm

import (
	"bytes"
	"time"

	"golang.org/x/exp/slog"
)

type HeaderFunc func(*bytes.Buffer, slog.Record) error

func DefaultHeader(w *bytes.Buffer, r slog.Record) error {
	fmtTs(w, r.Time, Tdefault)
	w.WriteByte(' ')
	w.WriteString(r.Level.String())
	w.WriteByte(' ')
	return nil
}

func itoa(buf *bytes.Buffer, i, w int) {
	if i < 0 {
		buf.WriteByte('-')
	}
	uitoa(buf, -i, w)
}

func uitoa(buf *bytes.Buffer, i, w int) {
	var tmp [20]byte
	wp := 19
	for i > 9 {
		q := i / 10
		tmp[wp] = byte('0' + i - 10*q)
		wp--
		w--
		i = q
	}
	tmp[wp] = byte('0' + i)
	for w > 1 {
		wp--
		tmp[wp] = '0'
		w--
	}
	buf.Write(tmp[wp:])
}

type timeFormat uint

const (
	Tdate = 1 << iota
	Tyear
	TUTC
	Tmillis
	Tmicros
)

const Tdefault = Tdate

func fmtTs(buf *bytes.Buffer, t time.Time, fmt timeFormat) {
	if fmt&TUTC != 0 {
		t = t.UTC()
	}
	if fmt&Tdate != 0 {
		ye, mo, dy := t.Date()
		if fmt&Tyear != 0 {
			uitoa(buf, ye, 4)
			buf.WriteByte(' ')
		}
		buf.WriteString(mo.String()[:3])
		buf.WriteByte(' ')
		uitoa(buf, dy, 2)
		buf.WriteByte(' ')
	}
	ho, mi, sc := t.Clock()
	uitoa(buf, ho, 2)
	buf.WriteByte(':')
	uitoa(buf, mi, 2)
	buf.WriteByte(':')
	uitoa(buf, sc, 2)
	if fmt&Tmicros != 0 {
		buf.WriteByte('.')
		uitoa(buf, t.Nanosecond()/1000, 6)
	} else if fmt&Tmillis != 0 {
		buf.WriteByte('.')
		uitoa(buf, t.Nanosecond()/1000000, 3)
	}
}
