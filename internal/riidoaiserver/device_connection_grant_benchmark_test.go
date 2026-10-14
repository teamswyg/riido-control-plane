package riidoaiserver

import (
	"strconv"
	"testing"
	"time"
)

func BenchmarkDeviceConnectionSummaryIsDeviceScoped(b *testing.B) {
	store := NewDevelopmentAIAgentClientStore()
	now := time.Unix(1, 0)
	for i := 0; i < 10_000; i++ {
		deviceID := "other-device-" + strconv.Itoa(i)
		store.upsertDeviceConnectionGrantLocked(deviceID, AuthorizationResult{
			PrincipalID: "other-account",
			WorkspaceID: "other-workspace",
		}, now)
	}
	for i := 0; i < 4; i++ {
		store.upsertDeviceConnectionGrantLocked("target-device", AuthorizationResult{
			PrincipalID: "account-" + strconv.Itoa(i),
			WorkspaceID: "workspace-" + strconv.Itoa(i),
		}, now)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.deviceConnectionSummaryLocked("target-device")
	}
}
