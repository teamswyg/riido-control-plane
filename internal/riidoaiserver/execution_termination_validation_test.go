package riidoaiserver

import (
	"strings"
	"testing"
)

func TestExecutionTerminationStrictBody(t *testing.T) {
	for _, body := range []string{`null`, `[]`, `{}`, strings.Replace(terminationFixtureBody, `:2}`, `:null}`, 1), strings.Replace(terminationFixtureBody, `:2}`, `:-1}`, 1), strings.Replace(terminationFixtureBody, `:2}`, `:1.5}`, 1), strings.Replace(terminationFixtureBody, `:2}`, `:18446744073709551616}`, 1), strings.Replace(terminationFixtureBody, `:2}`, `:2,"provider":"codex"}`, 1), strings.Replace(terminationFixtureBody, `:2}`, `:2,"started_invocations":3}`, 1), terminationFixtureBody + ` {}`, strings.Replace(terminationFixtureBody, "local_invocations_quiescent", "all_processes_exited", 1)} {
		t.Run(body, func(t *testing.T) {
			f := newTerminationHTTPFixture(t, StoreConfig{})
			w := serveTermination(f.handler, f.request(body))
			if w.Code != 400 {
				t.Fatalf("status=%d", w.Code)
			}
		})
	}
	f := newTerminationHTTPFixture(t, StoreConfig{})
	w := serveTermination(f.handler, f.request(terminationFixtureBody+strings.Repeat(" ", 4096)))
	if w.Code != 413 {
		t.Fatalf("oversize status=%d", w.Code)
	}
}
