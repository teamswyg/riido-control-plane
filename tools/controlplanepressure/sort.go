package main

import (
	"slices"
	"time"
)

func sortDurations(values []time.Duration) {
	slices.Sort(values)
}
