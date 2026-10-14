package riidoaiserver

func developmentSeedEvents(device DeviceRecord, daemon DeviceDaemonRecord) []ClientStreamEvent {
	return []ClientStreamEvent{
		{
			Seq:       1,
			EventType: AgentClientEventDeviceRuntimeSnapshot,
			Payload: DeviceRuntimeSnapshotEvent{
				EventType:     AgentClientEventDeviceRuntimeSnapshot,
				SchemaVersion: SchemaVersion,
				Device:        device,
			},
		},
		{
			Seq:       2,
			EventType: AgentClientEventDeviceDaemonStatus,
			Payload: DeviceDaemonStatusEvent{
				EventType:     AgentClientEventDeviceDaemonStatus,
				SchemaVersion: SchemaVersion,
				Daemon:        daemon,
			},
		},
		{
			Seq:       3,
			EventType: AgentClientEventWorkStatusChanged,
			Payload: AgentWorkStatusChangedEvent{
				EventType:       AgentClientEventWorkStatusChanged,
				SchemaVersion:   SchemaVersion,
				AgentID:         "agent-owned-codex",
				TaskID:          "task-1",
				ThreadID:        "thread-task-1-codex-2",
				RunID:           "run-dev-1",
				WorkStatus:      AgentWorkStatusQueued,
				AssignmentState: AgentAssignmentStateQueued,
				CommentKind:     AgentTaskCommentQueuedByBusyAgent,
			},
		},
	}
}
