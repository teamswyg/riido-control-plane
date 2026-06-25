package main

import (
	"runtime"
	"time"
)

type resourceSample struct {
	mem        runtime.MemStats
	goroutines int
}

func sampleResources() resourceSample {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	return resourceSample{mem: mem, goroutines: runtime.NumGoroutine()}
}

func diffResources(before, after resourceSample, ops int64) resourceDelta {
	delta := resourceDelta{
		HeapAllocBytes:  int64(after.mem.HeapAlloc) - int64(before.mem.HeapAlloc),
		TotalAllocBytes: int64(after.mem.TotalAlloc) - int64(before.mem.TotalAlloc),
		Mallocs:         int64(after.mem.Mallocs) - int64(before.mem.Mallocs),
		Frees:           int64(after.mem.Frees) - int64(before.mem.Frees),
		Goroutines:      after.goroutines - before.goroutines,
	}
	if ops > 0 {
		delta.TotalAllocPerOp = float64(delta.TotalAllocBytes) / float64(ops)
	}
	return delta
}

func summarize(values []time.Duration) latencySummary {
	return summarizeWithScale(values, time.Millisecond)
}

func summarizeMicros(values []time.Duration) latencySummary {
	return summarizeWithScale(values, time.Microsecond)
}

func summarizeWithScale(values []time.Duration, scale time.Duration) latencySummary {
	if len(values) == 0 {
		return latencySummary{}
	}
	sortDurations(values)
	return latencySummary{
		Min: scaledDuration(values[0], scale),
		P50: percentile(values, 0.50, scale),
		P90: percentile(values, 0.90, scale),
		P95: percentile(values, 0.95, scale),
		P99: percentile(values, 0.99, scale),
		Max: scaledDuration(values[len(values)-1], scale),
	}
}

func percentile(values []time.Duration, p float64, scale time.Duration) int64 {
	idx := int(float64(len(values)-1) * p)
	return scaledDuration(values[idx], scale)
}

func scaledDuration(value, scale time.Duration) int64 {
	return int64(value / scale)
}
