package main

import "testing"

func TestCronIntervalMinutesSupportsDailyAndHourly(t *testing.T) {
	cases := map[string]int{
		"17 20 * * *":  1440,
		"17 * * * *":   60,
		"*/30 * * * *": 30,
		"0 */6 * * *":  360,
	}
	for expr, want := range cases {
		got, err := cronIntervalMinutes(expr)
		if err != nil {
			t.Fatalf("%s: %v", expr, err)
		}
		if got != want {
			t.Fatalf("%s: got %d want %d", expr, got, want)
		}
	}
}

func TestCronIntervalMinutesRejectsWeeklyCadence(t *testing.T) {
	if _, err := cronIntervalMinutes("17 20 * * 1"); err == nil {
		t.Fatal("expected weekly cron to fail")
	}
}
