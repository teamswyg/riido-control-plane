package riidoaiserver

import (
	"encoding/hex"
	"strings"
)

func bootstrapDigest(value string) bool {
	b, err := hex.DecodeString(value)
	return err == nil && len(b) == 32 && strings.ToLower(value) == value
}

func bootstrapLegacyDevice(device string) bool {
	if !strings.HasPrefix(device, "dev_") || len(device) != 36 {
		return false
	}
	b, err := hex.DecodeString(device[4:])
	return err == nil && len(b) == 16 && strings.ToLower(device) == device
}

// A legacy ID remains stable. Only a server-issued, explicit bootstrap
// generation can make it covered; the ordinary enrollment path cannot mint it.
func coverageDeviceIdentity(device, generation string) bool {
	return strings.HasPrefix(device, executionCoveragePrefix) ||
		(bootstrapLegacyDevice(device) && strings.HasPrefix(generation, deviceBootstrapCoveragePrefix) && bootstrapDigest(strings.TrimPrefix(generation, deviceBootstrapCoveragePrefix)))
}
