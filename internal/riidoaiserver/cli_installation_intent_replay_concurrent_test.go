package riidoaiserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCLIInstallationIntentV2ConcurrentRequestsHaveOneOriginal(t *testing.T) {
	for _, durable := range []bool{false, true} {
		t.Run(map[bool]string{false: "local", true: "dynamo"}[durable], func(t *testing.T) {
			var config StoreConfig
			if durable {
				_, newStore, _ := newTerminationDynamoFixture(t)
				config.OperationStore = newStore()
			}
			f := newCLIInstallationFixture(t, config)
			body := installationIntentReplayBody(t, f)
			encoded, err := json.Marshal(body)
			if err != nil {
				t.Fatal(err)
			}
			responses := make(chan *httptest.ResponseRecorder, 8)
			for range cap(responses) {
				go func() {
					req := httptest.NewRequest(http.MethodPost, "/v1/internal/cli-installations/link-intents", bytes.NewReader(encoded))
					req.Header.Set("Authorization", "Bearer service-fixture-token")
					w := httptest.NewRecorder()
					f.handler.ServeHTTP(w, req)
					responses <- w
				}()
			}
			var original string
			for range cap(responses) {
				response := <-responses
				assertInstallationIntentReplayResponse(t, response, body["request_id"])
				if original != "" && original != response.Body.String() {
					t.Fatal("concurrent request published different originals")
				}
				original = response.Body.String()
			}
		})
	}
}
