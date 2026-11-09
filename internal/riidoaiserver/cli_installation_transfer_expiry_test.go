package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestCLIInstallationTransferRechecksExpiryAfterAuthorityRead(t *testing.T) {
	f := newCLITransferFixture(t)
	client := *f.operations.httpClient
	transport := client.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	expired := false
	client.Transport = transferTestTransport(func(req *http.Request) (*http.Response, error) {
		var body coverageLifecycleRequest
		if req.GetBody != nil {
			copyBody, err := req.GetBody()
			if err != nil {
				return nil, err
			}
			err = json.NewDecoder(copyBody).Decode(&body)
			_ = copyBody.Close()
			if err != nil {
				return nil, err
			}
		}
		response, err := transport.RoundTrip(req)
		if err == nil && req.Header.Get("X-Amz-Target") == dynamoDBGetItemTarget && body.Key["pk"]["S"] == "DEVICE_POLICY#"+f.target.DeviceID {
			expired = true
			f.operations.now = func() time.Time { return f.intent.ExpiresAt }
		}
		return response, err
	})
	f.operations.httpClient = &client
	status, _ := f.confirm(t)
	if !expired || status != http.StatusGone {
		t.Fatalf("expired during authority read status=%d", status)
	}
	f.assertOriginal(t)
}
