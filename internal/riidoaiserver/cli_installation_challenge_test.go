package riidoaiserver

import (
	"encoding/base64"
	"testing"
)

func mustDecodeCLIChallenge(t *testing.T, value string) []byte {
	t.Helper()
	data, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || len(data) != 32 {
		t.Fatal("challenge must be 32 random bytes")
	}
	return data
}
