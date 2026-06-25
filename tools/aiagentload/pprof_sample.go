package main

import (
	"context"
	"io"
	"net/http"
	"time"
)

func fetchPprofSample(ctx context.Context, client *http.Client, base string, target pprofTarget) pprofSample {
	start := time.Now()
	sample := pprofSample{Name: target.name, Path: target.path}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+target.path, nil)
	if err != nil {
		sample.ErrorCategory = errorCategory(err.Error())
		sample.LatencyMs = time.Since(start).Milliseconds()
		return sample
	}
	resp, err := client.Do(req)
	if err != nil {
		sample.ErrorCategory = errorCategory(err.Error())
		sample.LatencyMs = time.Since(start).Milliseconds()
		return sample
	}
	defer resp.Body.Close()
	bytes, err := io.Copy(io.Discard, resp.Body)
	sample.Status = resp.StatusCode
	sample.Bytes = bytes
	sample.LatencyMs = time.Since(start).Milliseconds()
	if err != nil {
		sample.ErrorCategory = errorCategory(err.Error())
	}
	return sample
}
