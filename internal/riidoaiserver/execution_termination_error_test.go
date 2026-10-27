package riidoaiserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"testing"
)

type terminationFailingDevice struct {
	DeviceCredentialStore
	fail       error
	panicValue string
}

func (d terminationFailingDevice) AuthorizeDeviceCredential(context.Context, string, string, AuthorizationRequest) (AuthorizationResult, error) {
	if d.panicValue != "" {
		panic(d.panicValue)
	}
	return AuthorizationResult{}, d.fail
}

func TestExecutionTerminationDependencyErrorsDoNotExposeSecrets(t *testing.T) {
	for _, tc := range []struct {
		name   string
		device terminationFailingDevice
		want   int
	}{{"error", terminationFailingDevice{fail: errors.New("secret-sentinel lease-sentinel")}, 503}, {"panic", terminationFailingDevice{panicValue: "secret-sentinel lease-sentinel"}, 503}, {"deadline", terminationFailingDevice{fail: context.DeadlineExceeded}, 504}, {"cancel", terminationFailingDevice{fail: context.Canceled}, 408}} {
		t.Run(tc.name, func(t *testing.T) {
			f := newTerminationHTTPFixture(t, StoreConfig{})
			recorder := &recordingTraceRecorder{}
			h := NewServer(ServerConfig{Assignment: f.actor, DeviceCredentials: tc.device, TraceRecorder: recorder}).Handler()
			var logs bytes.Buffer
			previous := log.Writer()
			log.SetOutput(&logs)
			defer log.SetOutput(previous)
			w := serveTermination(h, f.request(terminationFixtureBody))
			trace, err := json.Marshal(recorder.snapshot())
			if err != nil {
				t.Fatal(err)
			}
			if len(recorder.snapshot()) != 1 {
				t.Fatal("termination route did not produce bounded trace")
			}
			if w.Code != tc.want || strings.Contains(w.Body.String()+logs.String()+string(trace), "sentinel") || strings.Contains(logs.String()+string(trace), f.assignment.ID) {
				t.Fatalf("status=%d", w.Code)
			}
		})
	}
}
