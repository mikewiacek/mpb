package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/mikewiacek/mpb"
	"github.com/mikewiacek/mpb/decor"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	var wg sync.WaitGroup
	p := mpb.New(mpb.WithWaitGroup(&wg))
	total := 100
	numBars := 3
	wg.Add(numBars)

	for i := 0; i < numBars; i++ {
		name := fmt.Sprintf("Bar#%d:", i)
		b := p.AddBar(int64(total),
			mpb.BarOptOnCond(mpb.BarWidth(40), func() bool { return i > 0 }),
			mpb.PrependDecorators(
				decor.Name(name, decor.WCSyncWidth),
				decor.CountersNoUnit("%d / %d", decor.WCSyncSpace),
			),
			mpb.AppendDecorators(
				decor.EwmaETA(decor.ET_STYLE_GO, 60, decor.WC{W: 3}),
			),
		)
		go func() {
			defer wg.Done()
			max := 100 * time.Millisecond
			for i := 0; i < total; i++ {
				start := time.Now()
				time.Sleep(time.Duration(rand.Intn(10)+1) * max / 10)
				if i&1 == 1 {
					priority := total - int(b.Current())
					b.SetPriority(priority)
				}
				// ewma based decorators require work duration measurement
				b.IncrBy(1, time.Since(start))
			}
		}()
	}

	p.Wait()
}
