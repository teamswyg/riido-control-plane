package main

import (
	"runtime"
	"runtime/metrics"
	"syscall"
	"time"
)

const (
	cpuGCMetric       = "/cpu/classes/gc/total:cpu-seconds"
	cpuScavengeMetric = "/cpu/classes/scavenge/total:cpu-seconds"
)

type cpuSample struct {
	at       time.Time
	user     float64
	system   float64
	gc       float64
	scavenge float64
}

func sampleCPU() cpuSample {
	var usage syscall.Rusage
	_ = syscall.Getrusage(syscall.RUSAGE_SELF, &usage)
	samples := []metrics.Sample{
		{Name: cpuGCMetric},
		{Name: cpuScavengeMetric},
	}
	metrics.Read(samples)
	return cpuSample{
		at:       time.Now(),
		user:     timevalSeconds(usage.Utime),
		system:   timevalSeconds(usage.Stime),
		gc:       metricFloat(samples[0]),
		scavenge: metricFloat(samples[1]),
	}
}

func metricFloat(sample metrics.Sample) float64 {
	if sample.Value.Kind() != metrics.KindFloat64 {
		return 0
	}
	return sample.Value.Float64()
}

func addCPUDelta(delta *resourceDelta, before, after cpuSample) {
	delta.UserCPUSeconds = nonNegative(after.user - before.user)
	delta.SystemCPUSeconds = nonNegative(after.system - before.system)
	delta.GCCPUSeconds = nonNegative(after.gc - before.gc)
	delta.ScavengeCPUSeconds = nonNegative(after.scavenge - before.scavenge)
	delta.AvailableCPUSeconds = availableCPUSeconds(before.at, after.at)
	delta.ActiveCPUSeconds = delta.UserCPUSeconds + delta.SystemCPUSeconds
	if delta.AvailableCPUSeconds > 0 {
		delta.CPUUtilizationPct = delta.ActiveCPUSeconds / delta.AvailableCPUSeconds * 100
	}
}

func timevalSeconds(value syscall.Timeval) float64 {
	return float64(value.Sec) + float64(value.Usec)/1_000_000
}

func availableCPUSeconds(before, after time.Time) float64 {
	if after.Before(before) {
		return 0
	}
	return after.Sub(before).Seconds() * float64(runtime.GOMAXPROCS(0))
}

func nonNegative(value float64) float64 {
	if value < 0 {
		return 0
	}
	return value
}
