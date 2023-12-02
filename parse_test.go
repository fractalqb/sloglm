package sloglm

// slogite

import (
	"fmt"
	"testing"
	"time"

	"git.fractalqb.de/fractalqb/sllm/v3"
)

func ExampleParseTS() {
	tref := time.Date(2006, time.November, 12, 0, 0, 0, 0, time.UTC)
	t, n, err := ParseTS([]byte("2006-11-12 Mo 15:24:35.987654-07"), tref)
	fmt.Println(n, t, err)
	// Output:
	// 32 2006-11-12 15:24:35.987654 -0700 -0700 <nil>
}

func TestParseTS(t *testing.T) {
	ts := time.Date(2023, time.April, 1, 13, 42, 18, 432123000, time.UTC)
	var buf []byte
	tfmt := sllm.TimeFormat(0)
	defer func() {
		if p := recover(); p != nil {
			t.Fatalf("panic tfmt=%[1]b(%[1]d) ts='%s': %v", tfmt, string(buf), p)
		}
	}()
	const end = sllm.TMicros << 1
	for ; tfmt < end; tfmt++ {
		buf = buf[:0]
		buf = tfmt.Fmt(ts).AppendSllm(buf)
		if len(buf) == 0 {
			continue
		}
		us, _, err := ParseTS(buf, ts)
		if err != nil {
			t.Errorf("'%s': %s", string(buf), err)
		} else {
			expect := ts
			if tfmt&sllm.TMicros != 0 {
				expect = expect.Round(time.Microsecond)
				us = us.Round(time.Microsecond)
			} else if tfmt&sllm.TMillis != 0 {
				expect = expect.Round(time.Millisecond)
				us = us.Round(time.Millisecond)
			} else {
				expect = expect.Round(time.Second)
				us = us.Round(time.Second)
			}
			if !us.Equal(expect) {
				t.Errorf("%[1]b(%[1]d) %s =/= %s", tfmt, us, expect)
			}
		}
	}
}

func FuzzParseTS(f *testing.F) {
	f.Add("11-12")
	f.Add("2006-11-12")
	f.Add("11-12 Mo")
	f.Add("2006-11-12 Mo")
	f.Add("15:24:35")
	f.Add("15:24:35.123")
	f.Add("15:24:35.123456")
	f.Add("15:24:35+01")
	f.Add("15:24:35-01")
	f.Add("11-12 15:24:35")
	f.Add("2006-11-12 Mo 15:24:35.987654-07")
	tref := time.Date(2006, time.November, 12, 15, 24, 35, 987654000, time.UTC)
	f.Fuzz(func(t *testing.T, ts string) {
		defer func() {
			if p := recover(); p != nil {
				t.Errorf("'%s' -> %+v", ts, p)
			}
		}()
		ParseTSString(ts, tref)
	})
}

// func TestParseTS_fuzzfail(t *testing.T) {
// 	tref := time.Date(2006, time.November, 12, 15, 24, 35, 987654000, time.UTC)
// 	_, _, err := ParseTSString("00:00000.", tref)
// 	if err == nil {
// 		t.Fatal("fixed")
// 	}
// }

func BenchmarkParseTSString(b *testing.B) {
	tref := time.Date(2006, time.November, 12, 0, 0, 0, 0, time.UTC)
	for i := 0; i < b.N; i++ {
		ParseTSString("2006-11-12 Mo 15:24:35.999999-07", tref)
	}
}

func BenchmarkParseTS(b *testing.B) {
	tref := time.Date(2006, time.November, 12, 0, 0, 0, 0, time.UTC)
	for i := 0; i < b.N; i++ {
		ParseTS([]byte("2006-11-12 Mo 15:24:35.999999-07"), tref)
	}
}

func BenchmarkParseTime(b *testing.B) {
	for i := 0; i < b.N; i++ {
		time.Parse(time.RFC1123Z, "Mon, 12 Nov 2006 15:24:35 -0700")
	}
}
