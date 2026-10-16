package riidoaiserver

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"strconv"
	"time"
)

func authorizationJWTPayloadHints(out authorizationTokenObservation, encoded string, now time.Time) authorizationTokenObservation {
	out.format, out.expiryHint = "malformed_jwt", "invalid"
	payload, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return out
	}
	dec := json.NewDecoder(bytes.NewReader(payload))
	opening, err := dec.Token()
	if err != nil || opening != json.Delim('{') {
		return out
	}
	var expiry json.RawMessage
	for fields := 0; dec.More(); fields++ {
		if fields >= 64 {
			return out
		}
		rawKey, err := dec.Token()
		if err != nil {
			return out
		}
		key, ok := rawKey.(string)
		if !ok {
			return out
		}
		var value json.RawMessage
		if dec.Decode(&value) != nil {
			return out
		}
		if key == "exp" {
			if expiry != nil {
				return out
			}
			expiry = value
		}
	}
	closing, err := dec.Token()
	if err != nil || closing != json.Delim('}') || dec.Decode(&struct{}{}) != io.EOF {
		return out
	}
	out.format = "jwt"
	if expiry == nil {
		out.expiryHint = "absent"
		return out
	}
	seconds, err := strconv.ParseInt(string(expiry), 10, 64)
	if err != nil {
		return out
	}
	out.expiryHint = "future_unverified"
	if seconds <= now.Unix() {
		out.expiryHint = "past_unverified"
	}
	return out
}
