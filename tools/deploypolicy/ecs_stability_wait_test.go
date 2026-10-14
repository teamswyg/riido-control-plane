package deploypolicy

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestECSStabilityWaitExtendsOnlyAnActiveRollout(t *testing.T) {
	raw, err := os.ReadFile("../../.github/workflows/deploy-ai-agent-testnet.yml")
	if err != nil {
		t.Fatal(err)
	}
	script := string(raw)
	start := strings.Index(script, "# BEGIN bounded ECS stability wait")
	end := strings.Index(script, "# END bounded ECS stability wait")
	if start < 0 || end <= start {
		t.Fatal("bounded stability wait missing")
	}
	script = "set -euo pipefail\n" + script[start:end]
	for _, tc := range []struct {
		name, state        string
		first, final, want int
	}{
		{"initial success", "IN_PROGRESS", 0, 0, 1},
		{"active rollout settles", "IN_PROGRESS", 255, 0, 2},
		{"completion race confirms stability", "COMPLETED", 255, 0, 2},
		{"failed rollout cannot extend", "FAILED", 255, 0, 1},
		{"unknown rollout cannot extend", "None", 255, 0, 1},
		{"second wait still fails", "IN_PROGRESS", 255, 255, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			fake := `#!/usr/bin/env bash
set -euo pipefail
if [ "$2" = "wait" ]; then
 n=0; if [ -f "$COUNTER" ]; then n=$(cat "$COUNTER"); fi
 n=$((n+1)); echo "$n" > "$COUNTER"
 if [ "$n" = 1 ]; then exit "$FIRST"; fi
 exit "$FINAL"
fi
if [ "$2" = "describe-services" ]; then printf '%s\n' "$STATE"; exit 0; fi
exit 99
`
			if err := os.WriteFile(filepath.Join(dir, "aws"), []byte(fake), 0o700); err != nil {
				t.Fatal(err)
			}
			counter := filepath.Join(dir, "count")
			cmd := exec.Command("bash", "-c", script)
			cmd.Env = append(os.Environ(), "PATH="+dir+":"+os.Getenv("PATH"), "COUNTER="+counter, "ECS_CLUSTER=fixture", "ECS_SERVICE=fixture", "STATE="+tc.state, "FIRST="+strconv.Itoa(tc.first), "FINAL="+strconv.Itoa(tc.final))
			output, err := cmd.CombinedOutput()
			wantSuccess := tc.first == 0 || ((tc.state == "IN_PROGRESS" || tc.state == "COMPLETED") && tc.final == 0)
			if (err == nil) != wantSuccess {
				t.Fatalf("result=%v output=%s", err, output)
			}
			n, err := os.ReadFile(counter)
			if err != nil {
				t.Fatal(err)
			}
			if strings.TrimSpace(string(n)) != strconv.Itoa(tc.want) {
				t.Fatalf("wait count=%s", n)
			}
		})
	}
}
