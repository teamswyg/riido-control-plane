package main

import "time"

func perSecond(count int, duration time.Duration) float64 {
	if duration <= 0 {
		return 0
	}
	return float64(count) / duration.Seconds()
}

func failureRatePct(total, failures int) float64 {
	if total == 0 {
		return 0
	}
	return float64(failures) * 100 / float64(total)
}
