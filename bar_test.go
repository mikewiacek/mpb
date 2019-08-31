package mpb_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"unicode/utf8"

	. "github.com/mikewiacek/mpb"
	"github.com/mikewiacek/mpb/decor"
)

func TestBarCompleted(t *testing.T) {
	p := New(WithOutput(ioutil.Discard))
	total := 80
	bar := p.AddBar(int64(total))

	var count int
	for !bar.Completed() {
		time.Sleep(10 * time.Millisecond)
		bar.Increment()
		count++
	}

	p.Wait()
	if count != total {
		t.Errorf("got count: %d, expected %d\n", count, total)
	}
}

func TestBarID(t *testing.T) {
	p := New(WithOutput(ioutil.Discard))
	total := 100
	wantID := 11
	bar := p.AddBar(int64(total), BarID(wantID))

	go func(total int) {
		for i := 0; i < total; i++ {
			time.Sleep(50 * time.Millisecond)
			bar.Increment()
		}
	}(total)

	gotID := bar.ID()
	if gotID != wantID {
		t.Errorf("Expected bar id: %d, got %d\n", wantID, gotID)
	}

	bar.Abort(true)
	p.Wait()
}

func TestBarSetRefill(t *testing.T) {
	var buf bytes.Buffer

	width := 100
	p := New(WithOutput(&buf), WithWidth(width))

	total := 100
	till := 30
	refillRune, _ := utf8.DecodeLastRuneInString(DefaultBarStyle)

	bar := p.AddBar(int64(total), TrimSpace())

	bar.SetRefill(int64(till))
	bar.IncrBy(till)

	for i := 0; i < total-till; i++ {
		bar.Increment()
		time.Sleep(10 * time.Millisecond)
	}

	p.Wait()

	wantBar := fmt.Sprintf("[%s%s]",
		strings.Repeat(string(refillRune), till-1),
		strings.Repeat("=", total-till-1),
	)

	got := string(getLastLine(buf.Bytes()))

	if !strings.Contains(got, wantBar) {
		t.Errorf("Want bar: %q, got bar: %q\n", wantBar, got)
	}
}

func TestBarHas100PercentWithOnCompleteDecorator(t *testing.T) {
	var buf bytes.Buffer

	p := New(WithOutput(&buf))

	total := 50

	bar := p.AddBar(int64(total),
		AppendDecorators(
			decor.OnComplete(
				decor.Percentage(), "done",
			),
		),
	)

	for i := 0; i < total; i++ {
		bar.Increment()
		time.Sleep(10 * time.Millisecond)
	}

	p.Wait()

	hundred := "100 %"
	if !bytes.Contains(buf.Bytes(), []byte(hundred)) {
		t.Errorf("Bar's buffer does not contain: %q\n", hundred)
	}
}

func TestBarHas100PercentWithBarRemoveOnComplete(t *testing.T) {
	var buf bytes.Buffer

	p := New(WithOutput(&buf))

	total := 50

	bar := p.AddBar(int64(total),
		BarRemoveOnComplete(),
		AppendDecorators(decor.Percentage()),
	)

	for i := 0; i < total; i++ {
		bar.Increment()
		time.Sleep(10 * time.Millisecond)
	}

	p.Wait()

	hundred := "100 %"
	if !bytes.Contains(buf.Bytes(), []byte(hundred)) {
		t.Errorf("Bar's buffer does not contain: %q\n", hundred)
	}
}

func TestBarStyle(t *testing.T) {
	var buf bytes.Buffer
	customFormat := "╢▌▌░╟"
	p := New(WithOutput(&buf))
	total := 80
	bar := p.AddBar(int64(total), BarStyle(customFormat), TrimSpace())

	for i := 0; i < total; i++ {
		bar.Increment()
		time.Sleep(10 * time.Millisecond)
	}

	p.Wait()

	runes := []rune(customFormat)
	wantBar := fmt.Sprintf("%s%s%s",
		string(runes[0]),
		strings.Repeat(string(runes[1]), total-2),
		string(runes[len(runes)-1]),
	)
	got := string(getLastLine(buf.Bytes()))

	if !strings.Contains(got, wantBar) {
		t.Errorf("Want bar: %q:%d, got bar: %q:%d\n", wantBar, utf8.RuneCountInString(wantBar), got, utf8.RuneCountInString(got))
	}
}

func TestBarPanicBeforeComplete(t *testing.T) {
	var buf bytes.Buffer
	p := New(WithDebugOutput(&buf), WithOutput(ioutil.Discard))

	total := 100
	panicMsg := "Upps!!!"
	var pCount uint32
	bar := p.AddBar(int64(total),
		PrependDecorators(panicDecorator(panicMsg,
			func(st *decor.Statistics) bool {
				if st.Current >= 42 {
					atomic.AddUint32(&pCount, 1)
					return true
				}
				return false
			},
		)),
	)

	for i := 0; i < total; i++ {
		time.Sleep(10 * time.Millisecond)
		bar.Increment()
	}

	p.Wait()

	if pCount != 1 {
		t.Errorf("Decor called after panic %d times\n", pCount-1)
	}

	barStr := buf.String()
	if !strings.Contains(barStr, panicMsg) {
		t.Errorf("%q doesn't contain %q\n", barStr, panicMsg)
	}
}

func TestBarPanicAfterComplete(t *testing.T) {
	var buf bytes.Buffer
	p := New(WithDebugOutput(&buf), WithOutput(ioutil.Discard))

	total := 100
	panicMsg := "Upps!!!"
	var pCount uint32
	bar := p.AddBar(int64(total),
		PrependDecorators(panicDecorator(panicMsg,
			func(st *decor.Statistics) bool {
				if st.Completed {
					atomic.AddUint32(&pCount, 1)
					return true
				}
				return false
			},
		)),
	)

	for i := 0; i < total; i++ {
		time.Sleep(10 * time.Millisecond)
		bar.Increment()
	}

	p.Wait()

	if pCount != 1 {
		t.Errorf("Decor called after panic %d times\n", pCount-1)
	}

	barStr := buf.String()
	if !strings.Contains(barStr, panicMsg) {
		t.Errorf("%q doesn't contain %q\n", barStr, panicMsg)
	}
}

func panicDecorator(panicMsg string, cond func(*decor.Statistics) bool) decor.Decorator {
	d := &decorator{
		panicMsg: panicMsg,
		cond:     cond,
	}
	d.Init()
	return d
}

type decorator struct {
	decor.WC
	panicMsg string
	cond     func(*decor.Statistics) bool
}

func (d *decorator) Decor(st *decor.Statistics) string {
	if d.cond(st) {
		panic(d.panicMsg)
	}
	return d.FormatMsg("")
}
