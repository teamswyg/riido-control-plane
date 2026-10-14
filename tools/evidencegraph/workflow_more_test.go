package main

import "testing"

func TestWorkflowArtifactAndCadenceFailures(t *testing.T) {
	t.Parallel()
	upload := "uses: actions/upload-artifact@v4\nwith:\n  name: evidence\n  if-no-files-found: error\n"
	if !workflowUploadsStrictArtifact(upload, "evidence") {
		t.Fatalf("strict artifact should be accepted")
	}
	for _, artifact := range []string{"", "bad/name", "missing"} {
		if workflowUploadsStrictArtifact(upload, artifact) {
			t.Fatalf("artifact %q should fail", artifact)
		}
	}
	for _, text := range []string{
		"",
		"- cron: '* * *'",
		"- cron: '0 0 1 * *'",
		"- cron: '* * * * *'",
	} {
		if _, err := refreshCadenceMinutes(text); err == nil {
			t.Fatalf("cadence text %q should fail", text)
		}
	}
}

func TestMinuteHourIntervalBranches(t *testing.T) {
	t.Parallel()
	cases := []struct {
		minute string
		hour   string
		want   int
	}{
		{"0", "*/2", 120},
		{"30", "1", 24 * 60},
	}
	for _, tc := range cases {
		got, err := minuteHourInterval("expr", tc.minute, tc.hour)
		if err != nil {
			t.Fatalf("minuteHourInterval: %v", err)
		}
		if got != tc.want {
			t.Fatalf("minuteHourInterval = %d, want %d", got, tc.want)
		}
	}
	if _, err := minuteHourInterval("expr", "*", "*"); err == nil {
		t.Fatalf("wildcard minute/hour should fail")
	}
}
