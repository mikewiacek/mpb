package main

import (
	"fmt"
	"io"
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
	total, numBars := 100, 3
	wg.Add(numBars)

	for i := 0; i < numBars; i++ {
		name := fmt.Sprintf("Bar#%d:", i)
		efn := func(w io.Writer, width int, s *decor.Statistics) {
			if s.Completed {
				fmt.Fprintf(w, "Bar id: %d has been completed\n", s.ID)
			}
		}
		bar := p.AddBar(int64(total), mpb.BarExtender(mpb.FillerFunc(efn)),
			mpb.PrependDecorators(
				// simple name decorator
				decor.Name(name),
				// decor.DSyncWidth bit enables column width synchronization
				decor.Percentage(decor.WCSyncSpace),
			),
			mpb.AppendDecorators(
				// replace ETA decorator with "done" message, OnComplete event
				decor.OnComplete(
					// ETA decorator with ewma age of 60
					decor.EwmaETA(decor.ET_STYLE_GO, 60), "done",
				),
			),
		)
		// simulating some work
		go func() {
			defer wg.Done()
			max := 100 * time.Millisecond
			for i := 0; i < total; i++ {
				start := time.Now()
				time.Sleep(time.Duration(rand.Intn(10)+1) * max / 10)
				// ewma based decorators require work duration measurement
				bar.IncrBy(1, time.Since(start))
			}
		}()
	}
	// wait for all bars to complete and flush
	p.Wait()
}
