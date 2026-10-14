package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// The external boundary is Dynamo HTTP. Both application replicas use their
// native adapters and independent actor/client stores against this keyed table.
type membershipSharedDynamoFixture struct {
	mu                        sync.Mutex
	items                     map[string]map[string]map[string]string
	server                    *httptest.Server
	beforeScopedCommit        func(AIAgentClientSnapshot)
	beforeReceiptQuery        func()
	beforeOwnedSnapshotRead   func()
	rejectScopedPublication   bool
	ambiguousScopedCommit     bool
	beforePublicationRecovery func()
}

type membershipDynamoAction struct {
	TableName                 string                       `json:"TableName"`
	Key                       map[string]map[string]string `json:"Key"`
	Item                      map[string]map[string]string `json:"Item"`
	ConditionExpression       string                       `json:"ConditionExpression"`
	ExpressionAttributeValues map[string]map[string]string `json:"ExpressionAttributeValues"`
}

func newMembershipSharedDynamoFixture(t *testing.T) *membershipSharedDynamoFixture {
	t.Helper()
	f := &membershipSharedDynamoFixture{items: map[string]map[string]map[string]string{}}
	f.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.handle(t, w, r)
	}))
	t.Cleanup(f.server.Close)
	return f
}

func membershipDynamoKey(item map[string]map[string]string) string {
	return item["pk"]["S"] + "\x00" + item["sk"]["S"]
}
