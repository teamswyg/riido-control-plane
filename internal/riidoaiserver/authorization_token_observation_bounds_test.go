package riidoaiserver

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestAuthorizationTokenDiagnosticsBoundUntrustedInput(t *testing.T) {
	now := time.Unix(100, 0)
	jwt := func(payload string) string {
		return "e30." + base64.RawURLEncoding.EncodeToString([]byte(payload)) + ".signature"
	}
	manyFields := func(n int) string {
		fields := make([]string, n)
		for i := range fields {
			fields[i] = fmt.Sprintf(`"field%d":0`, i)
		}
		return "{" + strings.Join(fields, ",") + "}"
	}
	cases := []struct{ name, agent, bearer, source, format, expiry, relation string }{
		{"missing", "", "", "missing", "missing", "not_checked", "none"},
		{"opaque", "synthetic", "", "agent_header", "opaque", "not_checked", "agent_only"},
		{"past", jwt(`{"exp":99}`), "", "agent_header", "jwt", "past_unverified", "agent_only"},
		{"boundary", jwt(`{"exp":100}`), "", "agent_header", "jwt", "past_unverified", "agent_only"},
		{"future", jwt(`{"exp":101}`), "", "agent_header", "jwt", "future_unverified", "agent_only"},
		{"absent", jwt(`{"sub":"synthetic"}`), "", "agent_header", "jwt", "absent", "agent_only"},
		{"null", jwt(`{"exp":null}`), "", "agent_header", "jwt", "invalid", "agent_only"},
		{"string", jwt(`{"exp":"101"}`), "", "agent_header", "jwt", "invalid", "agent_only"},
		{"fraction", jwt(`{"exp":101.5}`), "", "agent_header", "jwt", "invalid", "agent_only"},
		{"duplicate", jwt(`{"exp":99,"exp":101}`), "", "agent_header", "malformed_jwt", "invalid", "agent_only"},
		{"array", jwt(`[101]`), "", "agent_header", "malformed_jwt", "invalid", "agent_only"},
		{"trailing", jwt(`{"exp":101}{}`), "", "agent_header", "malformed_jwt", "invalid", "agent_only"},
		{"bad-base64", "e30.!.signature", "", "agent_header", "malformed_jwt", "invalid", "agent_only"},
		{"token-bound", strings.Repeat("x", 8193), "", "agent_header", "oversized", "not_checked", "agent_only"},
		{"payload-bound", "e30." + strings.Repeat("a", 4097) + ".signature", "", "agent_header", "oversized", "not_checked", "agent_only"},
		{"fields64", jwt(manyFields(64)), "", "agent_header", "jwt", "absent", "agent_only"},
		{"fields65", jwt(manyFields(65)), "", "agent_header", "malformed_jwt", "invalid", "agent_only"},
		{"bearer", "", jwt(`{"exp":101}`), "bearer_header", "jwt", "future_unverified", "bearer_only"},
		{"both-same", "synthetic", "synthetic", "agent_header", "opaque", "not_checked", "both_same"},
		{"agent-priority", jwt(`{"exp":99}`), jwt(`{"exp":101}`), "agent_header", "jwt", "past_unverified", "both_different"},
		{"both-bound", strings.Repeat("x", 8193), strings.Repeat("x", 8193), "agent_header", "oversized", "not_checked", "both_not_compared"},
		{"bearer-bound-with-small-agent", "synthetic", strings.Repeat("x", 8193), "agent_header", "opaque", "not_checked", "both_not_compared"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.agent != "" {
				req.Header.Set(aiAgentTokenHeader, tc.agent)
			}
			if tc.bearer != "" {
				req.Header.Set("Authorization", "Bearer "+tc.bearer)
			}
			got := authorizationTokenHints(req, now)
			want := authorizationTokenObservation{source: tc.source, format: tc.format, expiryHint: tc.expiry, headerRelation: tc.relation}
			if got != want {
				t.Fatalf("finite hints got %#v, want %#v", got, want)
			}
		})
	}
}
