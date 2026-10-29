package riidoaiserver

import (
	"crypto/sha256"
	"encoding/binary"
)

func installationIntentRequestDigest(domain string, request cliInstallationIntentRequest, requestID string) [32]byte {
	return installationIntentTupleDigest(domain, cliInstallationServicePrincipal, request.Resource, request.ClientID, requestID)
}

func installationIntentTupleDigest(parts ...string) [32]byte {
	var canonical []byte
	for _, part := range parts {
		canonical = binary.BigEndian.AppendUint64(canonical, uint64(len(part)))
		canonical = append(canonical, part...)
	}
	return sha256.Sum256(canonical)
}
