package riidoaiserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
)

func TestCLIInstallationTransferRacingClaimPreservesOriginalLink(t *testing.T) {
	f := newCLITransferFixture(t)
	client := *f.operations.httpClient
	transport := client.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	raced := false
	client.Transport = transferTestTransport(func(req *http.Request) (*http.Response, error) {
		if !raced && req.Header.Get("X-Amz-Target") == dynamoDBTransactWriteTarget {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			req.Body = io.NopCloser(bytes.NewReader(body))
			var tx coverageLifecycleRequest
			if json.Unmarshal(body, &tx) != nil {
				return nil, errors.New("invalid transaction")
			}
			if len(tx.TransactItems) == 4 && tx.TransactItems[2].Put != nil {
				raced = true
				f.dynamo.mu.Lock()
				item := f.dynamo.items[membershipDynamoKey(devicePolicyKey(f.source.DeviceID))]
				item["revision"], item["unresolved"] = map[string]string{"N": "99"}, map[string]string{"N": "1"}
				f.dynamo.mu.Unlock()
			}
		}
		return transport.RoundTrip(req)
	})
	f.operations.httpClient = &client
	status, _ := f.confirm(t)
	if !raced || status != http.StatusServiceUnavailable {
		t.Fatalf("racing claim status=%d", status)
	}
	f.assertOriginal(t)
	intent, _, err := f.operations.loadCLIInstallationIntent(context.Background(), f.intent.IntentID)
	if err != nil || intent.Consumed {
		t.Fatal("failed transfer consumed approval")
	}
}
