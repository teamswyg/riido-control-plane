package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"testing"
	"time"
)

type transferTestTransport func(*http.Request) (*http.Response, error)

func (fn transferTestTransport) RoundTrip(req *http.Request) (*http.Response, error) { return fn(req) }

func TestCLIInstallationTransferLostACKReplaysCommittedRevision(t *testing.T) {
	f := newCLITransferFixture(t)
	client := *f.operations.httpClient
	transport := client.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	dropped := false
	client.Transport = transferTestTransport(func(req *http.Request) (*http.Response, error) {
		response, err := transport.RoundTrip(req)
		if !dropped && err == nil && req.Header.Get("X-Amz-Target") == dynamoDBTransactWriteTarget {
			dropped = true
			_, _ = io.Copy(io.Discard, response.Body)
			_ = response.Body.Close()
			return nil, errors.New("synthetic lost transfer acknowledgement")
		}
		return response, err
	})
	f.operations.httpClient = &client
	status, first := f.confirm(t)
	if !dropped || status != http.StatusOK || first.LinkRevision != "2" {
		t.Fatalf("lost ack replay status=%d", status)
	}
	f.operations.now = func() time.Time { return f.intent.ExpiresAt.Add(time.Hour) }
	status, replay := f.confirm(t)
	if status != http.StatusOK || !reflect.DeepEqual(first, replay) {
		t.Fatalf("consumed replay status=%d", status)
	}
	w := f.post(t, "/v1/internal/cli-installations/link-intents", f.wire, "service", "")
	var original cliInstallationIntentReplayResponse
	if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &original) != nil || !reflect.DeepEqual(original, f.intent) {
		t.Fatal("replay changed original challenge or expiry")
	}
	c, _, err := f.operations.loadExecutionCoverage(context.Background(), f.source.DeviceID)
	if err != nil || c.State != "revoked" || c.Unresolved != 0 {
		t.Fatal("source retirement lost")
	}
}
