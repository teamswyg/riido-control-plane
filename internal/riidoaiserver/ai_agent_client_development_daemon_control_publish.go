package riidoaiserver

func (s *DevelopmentAIAgentClientStore) publishDaemonControlLocked(daemon DeviceDaemonRecord, action DaemonControlAction) {
	s.putDaemonLocked(daemon)
	s.appendClientEventLocked(AgentClientEventDeviceDaemonStatus, DeviceDaemonStatusEvent{
		EventType:     AgentClientEventDeviceDaemonStatus,
		SchemaVersion: SchemaVersion,
		Daemon:        copyDeviceDaemon(daemon),
	})
	if action != DaemonControlActionStop {
		return
	}
	if device, ok := s.deviceByIDLocked(daemon.DeviceID); ok {
		s.appendClientEventLocked(AgentClientEventDeviceRuntimeSnapshot, DeviceRuntimeSnapshotEvent{
			EventType:     AgentClientEventDeviceRuntimeSnapshot,
			SchemaVersion: SchemaVersion,
			Device:        device,
		})
	}
}
