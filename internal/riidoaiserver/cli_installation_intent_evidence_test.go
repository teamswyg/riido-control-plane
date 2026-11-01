package riidoaiserver

import (
	"encoding/json"
	"maps"
	"net/http"
	"testing"
)

const (
	installationEvidencePath       = "/v1/internal/cli-installations/link-intent-evidence"
	installationEvidenceTestSchema = "riido-cli-installation-intent-evidence.v1"
)

func installationEvidenceBody(body map[string]string) map[string]string {
	result := maps.Clone(body)
	result["schema_version"] = installationEvidenceTestSchema
	return result
}

func TestCLIInstallationIntentEvidenceReadsWithoutCreating(t *testing.T) {
	for _, durable := range []bool{false, true} {
		t.Run(map[bool]string{false: "local", true: "dynamo"}[durable], func(t *testing.T) {
			var config StoreConfig
			var dynamo *terminationDynamoFixture
			if durable {
				var newStore func() *DynamoDBAssignmentOperationStore
				dynamo, newStore, _ = newTerminationDynamoFixture(t)
				config.OperationStore = newStore()
			}
			f := newCLIInstallationFixture(t, config)
			body := installationIntentReplayBody(t, f)
			for _, found := range []bool{false, true} {
				var original cliInstallationIntentResponse
				if found {
					original = assertInstallationIntentReplayResponse(t, f.post(t, "/v1/internal/cli-installations/link-intents", body, false), body["request_id"])
				}
				before := 0
				if dynamo != nil {
					dynamo.mu.Lock()
					before = len(dynamo.requests)
					dynamo.mu.Unlock()
				}
				response := f.post(t, installationEvidencePath, installationEvidenceBody(body), false)
				if dynamo != nil {
					assertEvidenceGetsOnly(t, dynamo, before)
				}
				var got struct {
					SchemaVersion string                               `json:"schema_version"`
					RequestID     string                               `json:"request_id"`
					Found         bool                                 `json:"found"`
					Intent        *cliInstallationIntentReplayResponse `json:"intent"`
				}
				if response.Code != http.StatusOK || json.Unmarshal(response.Body.Bytes(), &got) != nil {
					t.Fatalf("evidence status=%d", response.Code)
				}
				if got.SchemaVersion != installationEvidenceTestSchema || got.RequestID != body["request_id"] || got.Found != found || (got.Intent != nil) != found {
					t.Fatal("evidence changed request identity or presence")
				}
				if found && (got.Intent.RequestID != body["request_id"] || got.Intent.IntentID != original.IntentID || got.Intent.Challenge != original.Challenge || !got.Intent.ExpiresAt.Equal(original.ExpiresAt)) {
					t.Fatal("evidence did not preserve the immutable original response")
				}
			}
		})
	}
}
