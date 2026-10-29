package riidoaiserver

// preserveRuntimeSnapshotLivenessLocked keeps observations that have not reached
// the throttled durable snapshot yet. It merges timestamps, never registrations,
// credentials, commands, provider health, or account/workspace grants.
func (s *DevelopmentAIAgentClientStore) preserveRuntimeSnapshotLivenessLocked(devices []DeviceRecord, daemons map[string]DeviceDaemonRecord, credentials map[string]deviceCredentialRecord) {
	for i := range devices {
		device := &devices[i]
		previousCredential, hadCredential := s.deviceCredentials[device.DeviceID]
		credential, hasCredential := credentials[device.DeviceID]
		if !hadCredential || !hasCredential || previousCredential.secretHash != credential.secretHash ||
			previousCredential.ownerPrincipalID != credential.ownerPrincipalID || previousCredential.machineID != credential.machineID {
			continue
		}
		previous, ok := s.deviceByIDLocked(device.DeviceID)
		if !ok || previous.OwnerPrincipalID != device.OwnerPrincipalID {
			continue
		}
		for j := range device.Runtimes {
			runtime := &device.Runtimes[j]
			key := daemonStorageKey(DeviceDaemonRecord{DeviceID: device.DeviceID, DaemonID: runtime.DaemonID, Profile: runtime.DaemonProfile})
			observed, hadDaemon := s.daemons[key]
			durable, hasDaemon := daemons[key]
			if !hadDaemon || !hasDaemon || observed.OwnerPrincipalID != durable.OwnerPrincipalID ||
				!observed.StartedAt.Equal(durable.StartedAt) || observed.PID != durable.PID ||
				durable.Availability != DaemonAvailabilityOnline || observed.LastSeenAt.Before(durable.LastSeenAt) {
				continue
			}
			for _, oldRuntime := range previous.Runtimes {
				if oldRuntime.RuntimeID != runtime.RuntimeID || oldRuntime.DaemonID != runtime.DaemonID ||
					oldRuntime.DaemonProfile != runtime.DaemonProfile || oldRuntime.OwnerPrincipalID != runtime.OwnerPrincipalID {
					continue
				}
				if oldRuntime.LastDetectedAt.After(runtime.LastDetectedAt) {
					runtime.LastDetectedAt = oldRuntime.LastDetectedAt
				}
				if observed.LastSeenAt.After(device.DaemonLastSeenAt) {
					device.DaemonLastSeenAt = observed.LastSeenAt
				}
				// Keep durable commands and health; only these values are heartbeat observations.
				durable.LastSeenAt = observed.LastSeenAt
				durable.UptimeSeconds = observed.UptimeSeconds
				daemons[key] = durable
				break
			}
		}
	}
}
