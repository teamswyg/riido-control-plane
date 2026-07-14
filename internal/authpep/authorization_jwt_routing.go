package authpep

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strings"
)

// claimsConfiguredIssuer is the migration discriminator between Riido Auth
// access tokens and independently authenticated legacy JWTs. A token that
// claims this Auth issuer is owned by this PEP and can never fall through after
// a signature, claim, introspection, or domain-policy failure. Foreign or
// missing issuers remain available to the existing legacy authorizer.
func (a *JWTAccessTokenAuthorizer) claimsConfiguredIssuer(raw string) bool {
	if a == nil || raw == "" || len(raw) > maxJWTAccessTokenBytes || strings.TrimSpace(raw) != raw || strings.ContainsAny(raw, "\r\n\x00") {
		return false
	}
	segments := strings.Split(raw, ".")
	if len(segments) != 3 || segments[1] == "" {
		return false
	}
	payload, err := base64.RawURLEncoding.DecodeString(segments[1])
	if err != nil || len(payload) > 8<<10 {
		return false
	}
	issuer, found, containsConfiguredIssuer := topLevelJWTIssuer(payload, a.issuer)
	return containsConfiguredIssuer || (found && issuer == a.issuer)
}

func topLevelJWTIssuer(data []byte, configuredIssuer string) (issuer string, found, containsConfiguredIssuer bool) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	opening, err := decoder.Token()
	if err != nil || opening != json.Delim('{') {
		return "", false, false
	}
	seenIssuer := false
	for decoder.More() {
		token, err := decoder.Token()
		key, ok := token.(string)
		if err != nil || !ok {
			return issuer, found, containsConfiguredIssuer
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return issuer, found, containsConfiguredIssuer
		}
		if key != "iss" {
			continue
		}
		var candidate string
		if json.Unmarshal(value, &candidate) != nil {
			continue
		}
		if candidate == configuredIssuer {
			containsConfiguredIssuer = true
		}
		if !seenIssuer {
			issuer, found, seenIssuer = candidate, true, true
		}
	}
	return issuer, found, containsConfiguredIssuer
}
