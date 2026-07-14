package riidoaiserver

type DeviceAgentCapability string

const (
	DeviceAgentCapabilityReady          DeviceAgentCapability = "ready"
	DeviceAgentCapabilityOffline        DeviceAgentCapability = "offline"
	DeviceAgentCapabilityUpdateRequired DeviceAgentCapability = "update_required"
	DeviceAgentCapabilityVersionUnknown DeviceAgentCapability = "version_unknown"
)

type DeviceClientStatus struct {
	DesktopRegistered    bool                  `json:"desktop_registered"`
	DesktopAppVersion    string                `json:"desktop_app_version,omitempty"`
	DaemonAvailability   DaemonAvailability    `json:"daemon_availability"`
	DaemonVersion        string                `json:"daemon_version,omitempty"`
	MinimumDaemonVersion string                `json:"minimum_daemon_version"`
	LatestDaemonVersion  string                `json:"latest_daemon_version,omitempty"`
	AgentCapability      DeviceAgentCapability `json:"agent_capability"`
	AgentSupported       bool                  `json:"agent_supported"`
	UpdateRequired       bool                  `json:"update_required"`
	DownloadURL          string                `json:"download_url,omitempty"`
}

type DaemonClientCompatibilityPolicy struct {
	MinimumVersion string
	LatestVersion  string
	DownloadURL    string
}

func DefaultDaemonClientCompatibilityPolicy() DaemonClientCompatibilityPolicy {
	return DaemonClientCompatibilityPolicy{
		MinimumVersion: "v0.0.68",
		LatestVersion:  "v0.0.68",
		DownloadURL:    "https://cdn.riido.io/releases/latest/Riido-arm64.dmg",
	}
}
