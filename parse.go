package sloglm

import (
	"fmt"
	"time"
	"unsafe"
)

func ParseTSString(s string, tref time.Time) (t time.Time, n int, err error) {
	b := unsafe.Slice(unsafe.StringData(s), len(s))
	return ParseTS(b, tref)
}

func ParseTS(b []byte, tref time.Time) (t time.Time, n int, err error) {
	if len(b) < 5 {
		return t, 0, nil
	}
	var (
		year, day, hour, min, sec, nsec int
		month                           time.Month
	)
	if b[4] == '-' {
		if len(b) < 10 {
			return t, 0, nil
		}
		if year, err = parseUInt(b, 0, 4); err != nil {
			return t, 0, err
		}
		n = 5
	} else {
		year = tref.Year()
		if b[2] == ':' {
			n = -1
		}
	}
	if n < 0 {
		month = tref.Month()
		day = tref.Day()
		n = 0
	} else {
		mno, err := parseUInt(b, n, n+2)
		if err != nil {
			return t, n, err
		}
		month = time.Month(mno)
		if day, err = parseUInt(b, n+3, n+5); err != nil {
			return t, n + 3, err
		}
		if n+8 >= len(b) {
			return time.Date(year, month, day,
				tref.Hour(), tref.Minute(), tref.Second(), tref.Nanosecond(),
				tref.Location(),
			), n + 5, nil
		} else if b[n+8] == ' ' {
			n += 9
		} else {
			n += 6
		}
	}
	if len(b) >= n+8 {
		if b[n+2] != ':' || b[n+5] != ':' {
			hour, min, sec = tref.Clock()
		} else if hour, err = parseUInt(b, n, n+2); err != nil {
			hour, min, sec = tref.Clock()
		} else if min, err = parseUInt(b, n+3, n+5); err != nil {
			hour, min, sec = tref.Clock()
		} else if sec, err = parseUInt(b, n+6, n+8); err != nil {
			hour, min, sec = tref.Clock()
		}
		n += 8
	}
	if n+1 < len(b) && b[n] == '.' {
		n++
		d, p := b[n], 100000000
		for p > 0 && d >= '0' && d <= '9' {
			nsec += p * int(d-'0')
			p /= 10
			n++
			if n >= len(b) {
				break
			}
			d = b[n]
		}
	} else {
		hour, min, sec = tref.Clock()
		nsec = tref.Nanosecond()
	}
	var loc *time.Location
	if n < len(b) {
		switch b[n] {
		case '-':
			if n+2 < len(b) {
				dh, err := parseUInt(b, n+1, n+3)
				if err != nil {
					loc = tref.Location()
				} else {
					loc = time.FixedZone("", -dh*60*60)
				}
				n += 3
			} else {
				loc = tref.Location()
			}
		case '+':
			if n+2 < len(b) {
				dh, err := parseUInt(b, n+1, n+3)
				if err != nil {
					loc = tref.Location()
				} else {
					loc = time.FixedZone("", dh*60*60)
				}
				n += 3
			} else {
				loc = tref.Location()
			}
		default:
			loc = tref.Location()
		}
	} else {
		loc = tref.Location()
	}
	return time.Date(year, month, day, hour, min, sec, nsec, loc), n, nil
}

func digit(b byte) (int, error) {
	if b < '0' || b > '9' {
		return 0, fmt.Errorf("invalid digit '%c'", b)
	}
	return int(b - '0'), nil
}

func parseUInt(b []byte, s, e int) (i int, err error) {
	if i, err = digit(b[s]); err != nil {
		return 0, err
	}
	for s++; s < e; s++ {
		j, err := digit(b[s])
		if err != nil {
			return 0, err
		}
		i = 10*i + j
	}
	return i, nil
}
