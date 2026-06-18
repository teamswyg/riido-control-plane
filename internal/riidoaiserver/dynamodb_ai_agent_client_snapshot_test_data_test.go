package riidoaiserver

import "time"

func fixedSnapshotTestNow() time.Time {
	return time.Date(2026, 6, 2, 1, 2, 3, 0, time.UTC)
}

func snapshotTestRecord(savedAt time.Time) AIAgentClientSnapshot {
	return AIAgentClientSnapshot{
		SchemaVersion:           AIAgentClientPersistenceSchemaVersion,
		SavedAt:                 savedAt,
		WorkspaceID:             "workspace-dev",
		Devices:                 []DeviceRecord{{DeviceID: "device-a", OwnerPrincipalID: "user-1"}},
		Agents:                  []AgentClientRecord{{AgentID: "agent-a", OwnerPrincipalID: "user-1", WorkspaceID: "workspace-dev", Name: "Agent A", Visibility: AgentVisibilityPrivate}},
		TaskThreads:             map[string][]AIAgentTaskThreadRecord{},
		NextDeviceCredentialSeq: 1,
		NextDaemonCommand:       2,
	}
}

type snapshotQueryTestPayload struct {
	TableName              string                       `json:"TableName"`
	ConsistentRead         bool                         `json:"ConsistentRead"`
	KeyConditionExpression string                       `json:"KeyConditionExpression"`
	ExpressionValues       map[string]map[string]string `json:"ExpressionAttributeValues"`
}
