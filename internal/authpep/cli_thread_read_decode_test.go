package authpep

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestCLIThreadReadResponseStrictBoundary(t *testing.T) {
	a := &CLIThreadReadAuthorizer{resource: threadReadFixtureResource, now: time.Now}
	for _, tc := range []struct {
		name   string
		change func(*cliThreadReadResponse)
	}{
		{"false", func(r *cliThreadReadResponse) { r.Allowed = false }},
		{"schema", func(r *cliThreadReadResponse) { r.Schema = "other" }},
		{"principal", func(r *cliThreadReadResponse) { r.PrincipalID = "" }},
		{"resource", func(r *cliThreadReadResponse) { r.Resource = "https://other.test" }},
		{"purpose", func(r *cliThreadReadResponse) { r.Purpose = "write" }},
		{"workspace", func(r *cliThreadReadResponse) { r.WorkspaceID = "other" }},
		{"task", func(r *cliThreadReadResponse) { r.TaskID = "other" }},
		{"expired", func(r *cliThreadReadResponse) { r.ExpiresAt = time.Now().Add(-time.Second) }},
		{"long lived", func(r *cliThreadReadResponse) { r.ExpiresAt = r.CheckedAt.Add(time.Hour) }},
		{"stale verdict", func(r *cliThreadReadResponse) { r.CheckedAt = time.Now().Add(-6 * time.Second) }},
		{"future verdict", func(r *cliThreadReadResponse) { r.CheckedAt = time.Now().Add(6 * time.Second) }},
		{"unknown role", func(r *cliThreadReadResponse) { r.Roles = []string{"owner"} }},
		{"duplicate role", func(r *cliThreadReadResponse) { r.Roles = []string{"admin", "admin"} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := threadReadFixtureResponse()
			tc.change(&r)
			raw, _ := json.Marshal(r)
			if _, err := a.decode(raw, threadReadFixtureRequest()); err == nil {
				t.Fatal("invalid current authorization accepted")
			}
		})
	}
	good, _ := json.Marshal(threadReadFixtureResponse())
	for _, raw := range []string{
		strings.Replace(string(good), `"allowed":true`, `"allowed":false,"allowed":true`, 1),
		strings.Replace(string(good), `"allowed":true`, `"ALLOWED":false,"allowed":true`, 1),
		strings.Replace(string(good), `"allowed":true`, `"ALLOWED":true`, 1),
		strings.Replace(string(good), `"roles":[]`, `"roles":null`, 1),
		string(good) + `{}`,
	} {
		if _, err := a.decode([]byte(raw), threadReadFixtureRequest()); err == nil {
			t.Fatal("ambiguous response accepted")
		}
	}
	r := threadReadFixtureResponse()
	r.Roles = []string{"admin"}
	raw, _ := json.Marshal(r)
	if result, err := a.decode(raw, threadReadFixtureRequest()); err != nil || len(result.Roles) != 1 {
		t.Fatal("current admin role lost", err)
	}
}
