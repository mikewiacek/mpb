package mpb_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/mikewiacek/mpb"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestBarCount(t *testing.T) {
	p := mpb.New(mpb.WithOutput(ioutil.Discard))

	var wg sync.WaitGroup
	wg.Add(1)
	b := p.AddBar(100)
	go func() {
		for i := 0; i < 100; i++ {
			if i == 33 {
				wg.Done()
			}
			b.Increment()
			time.Sleep(randomDuration(100 * time.Millisecond))
		}
	}()

	wg.Wait()
	count := p.BarCount()
	if count != 1 {
		t.Errorf("BarCount want: %q, got: %q\n", 1, count)
	}

	b.Abort(true)
	p.Wait()
}

func TestBarAbort(t *testing.T) {
	p := mpb.New(mpb.WithOutput(ioutil.Discard))

	var wg sync.WaitGroup
	wg.Add(1)
	bars := make([]*mpb.Bar, 3)
	for i := 0; i < 3; i++ {
		b := p.AddBar(100)
		bars[i] = b
		go func(n int) {
			for i := 0; !b.Completed(); i++ {
				if n == 0 && i >= 33 {
					b.Abort(true)
					wg.Done()
				}
				b.Increment()
				time.Sleep(randomDuration(100 * time.Millisecond))
			}
		}(i)
	}

	wg.Wait()
	count := p.BarCount()
	if count != 2 {
		t.Errorf("BarCount want: %q, got: %q\n", 2, count)
	}
	bars[1].Abort(true)
	bars[2].Abort(true)
	p.Wait()
}

func TestWithContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	shutdown := make(chan struct{})
	p := mpb.NewWithContext(ctx,
		mpb.WithOutput(ioutil.Discard),
		mpb.WithRefreshRate(50*time.Millisecond),
		mpb.WithShutdownNotifier(shutdown),
	)

	total := 10000
	numBars := 3
	bars := make([]*mpb.Bar, 0, numBars)
	for i := 0; i < numBars; i++ {
		bar := p.AddBar(int64(total))
		bars = append(bars, bar)
		go func() {
			for !bar.Completed() {
				bar.Increment()
				time.Sleep(randomDuration(100 * time.Millisecond))
			}
		}()
	}

	time.Sleep(50 * time.Millisecond)
	cancel()

	p.Wait()
	select {
	case <-shutdown:
	case <-time.After(100 * time.Millisecond):
		t.Error("Progress didn't stop")
	}
}

func getLastLine(bb []byte) []byte {
	split := bytes.Split(bb, []byte("\n"))
	return split[len(split)-2]
}

func randomDuration(max time.Duration) time.Duration {
	return time.Duration(rand.Intn(10)+1) * max / 10
}
