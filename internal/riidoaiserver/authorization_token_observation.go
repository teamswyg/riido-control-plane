package riidoaiserver

import (
	"crypto/subtle"
	"net/http"
	"strings"
	"time"
)

type authorizationTokenObservation struct{ source, format, expiryHint, headerRelation string }

// These are unverified, bounded shape hints after denial. They never verify,
// authorize, renew, or select a different credential and never retain raw input.
func authorizationTokenHints(r *http.Request, now time.Time) authorizationTokenObservation {
	out := authorizationTokenObservation{source: "missing", format: "missing", expiryHint: "not_checked", headerRelation: "none"}
	agent := strings.TrimSpace(r.Header.Get(aiAgentTokenHeader))
	bearer, _ := requestBearerToken(r)
	switch {
	case agent != "" && bearer != "":
		out.headerRelation = "both_not_compared"
		if len(agent) <= 8192 && len(bearer) <= 8192 {
			out.headerRelation = "both_different"
			if subtle.ConstantTimeCompare([]byte(agent), []byte(bearer)) == 1 {
				out.headerRelation = "both_same"
			}
		}
	case agent != "":
		out.headerRelation = "agent_only"
	case bearer != "":
		out.headerRelation = "bearer_only"
	}
	token := agent
	switch {
	case token != "":
		out.source = "agent_header"
	case bearer != "":
		token, out.source = bearer, "bearer_header"
	default:
		return out
	}
	if len(token) > 8192 {
		out.format = "oversized"
		return out
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		out.format = "opaque"
		return out
	}
	if len(parts[1]) > 4096 {
		out.format = "oversized"
		return out
	}
	return authorizationJWTPayloadHints(out, parts[1], now)
}
