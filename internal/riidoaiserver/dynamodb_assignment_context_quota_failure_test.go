package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDynamoDBAssignmentContextQuotaBackendFailureNeverAllows(t *testing.T) {
	for _, body := range []string{`{"__type":"ProvisionedThroughputExceededException"}`, `{"__type":"AccessDeniedException"}`, `not-json`} {
		t.Run(body, func(t *testing.T) {
			endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(body))
			}))
			defer endpoint.Close()
			credentials, _ := NewStaticAWSCredentialsProvider("test", "test", "")
			store, err := NewDynamoDBAssignmentOperationStore(DynamoDBAssignmentOperationStoreConfig{Region: "test", TableName: "test", Endpoint: endpoint.URL, CredentialsProvider: credentials})
			if err != nil {
				t.Fatal(err)
			}
			defer store.Close()
			if ok, err := store.ConsumeAssignmentContextRead(t.Context(), "assignment"); ok || err == nil {
				t.Fatalf("failure allowed=%v err=%v", ok, err)
			}
		})
	}
}
