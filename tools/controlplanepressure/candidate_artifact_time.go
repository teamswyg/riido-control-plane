package main

import (
	"time"
)

const timeFormat = time.RFC3339

func hours(value int) time.Duration {
	return time.Duration(value) * time.Hour
}
