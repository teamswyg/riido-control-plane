package main

import "time"

const stalePartialAfterDays = 2

func readinessDate(value string) (time.Time, error) {
	return time.Parse("2006-01-02", value)
}

func partialAgeDays(date string, now time.Time) int {
	start, err := readinessDate(date)
	if err != nil {
		return 0
	}
	today := time.Date(now.UTC().Year(), now.UTC().Month(), now.UTC().Day(), 0, 0, 0, 0, time.UTC)
	if today.Before(start) {
		return 0
	}
	return int(today.Sub(start).Hours() / 24)
}
