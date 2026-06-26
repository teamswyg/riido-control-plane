package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestControlPlanePressureCanWriteLocalProfiles(t *testing.T) {
	root := t.TempDir()
	out := root + "/pressure.json"
	err := mainRun([]string{
		"-duration", "20ms",
		"-concurrency", "1",
		"-threads", "2",
		"-lines", "1",
		"-pprof-dir", root + "/pprof",
		"-evidence-out", out,
	})
	if err != nil {
		t.Fatal(err)
	}
	var got pressureReport
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	assertProfileArtifacts(t, got.Profiles)
}

func assertProfileArtifacts(t *testing.T, profiles []profileArtifact) {
	t.Helper()
	if len(profiles) != 3 {
		t.Fatalf("profiles = %+v", profiles)
	}
	for _, profile := range profiles {
		if profile.Kind == "" || profile.Path == "" || profile.Bytes <= 0 {
			t.Fatalf("invalid profile metadata: %+v", profile)
		}
	}
}
