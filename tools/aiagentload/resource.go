package main

import "runtime"

type resourceDelta struct {
	HeapAllocBytes       int64   `json:"heap_alloc_bytes"`
	TotalAllocBytes      int64   `json:"total_alloc_bytes"`
	TotalAllocPerRequest float64 `json:"total_alloc_bytes_per_request"`
	Mallocs              int64   `json:"mallocs"`
	Frees                int64   `json:"frees"`
	Goroutines           int     `json:"goroutines"`
}

type resourceSample struct {
	heapAlloc  uint64
	totalAlloc uint64
	mallocs    uint64
	frees      uint64
	goroutines int
}

func sampleResources() resourceSample {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return resourceSample{
		heapAlloc:  stats.HeapAlloc,
		totalAlloc: stats.TotalAlloc,
		mallocs:    stats.Mallocs,
		frees:      stats.Frees,
		goroutines: runtime.NumGoroutine(),
	}
}

func diffResources(before, after resourceSample, total int) resourceDelta {
	return resourceDelta{
		HeapAllocBytes:       int64(after.heapAlloc) - int64(before.heapAlloc),
		TotalAllocBytes:      int64(after.totalAlloc) - int64(before.totalAlloc),
		TotalAllocPerRequest: perRequest(after.totalAlloc-before.totalAlloc, total),
		Mallocs:              int64(after.mallocs) - int64(before.mallocs),
		Frees:                int64(after.frees) - int64(before.frees),
		Goroutines:           after.goroutines - before.goroutines,
	}
}

func perRequest(value uint64, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(value) / float64(total)
}
