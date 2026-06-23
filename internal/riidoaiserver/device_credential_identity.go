package riidoaiserver

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// deviceIDForMachine derives a stable DeviceID from the device's machine id.
// The DeviceID is not a secret; the rotating device secret is the auth factor.
func deviceIDForMachine(machineID string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(machineID)))
	return "dev_" + hex.EncodeToString(sum[:16])
}

func newDeviceSecret() (string, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate device secret: %w", err)
	}
	return "rdev_" + hex.EncodeToString(raw[:]), nil
}
