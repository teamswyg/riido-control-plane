package main

import (
	"sync"
	"sync/atomic"
	"time"
)

func worker(
	op func() error,
	deadline time.Time,
	out chan<- sample,
	ops *atomic.Int64,
	errors *atomic.Int64,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	local := int64(0)
	for time.Now().Before(deadline) {
		start := time.Now()
		err := op()
		local++
		ops.Add(1)
		if err != nil {
			errors.Add(1)
		}
		if local == 1 || local%64 == 0 {
			out <- sample{latency: time.Since(start)}
		}
	}
}
