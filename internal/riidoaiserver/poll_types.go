package riidoaiserver

type PollRequest struct {
	DaemonID  string `json:"daemon_id"`
	DeviceID  string `json:"device_id"`
	RuntimeID string `json:"runtime_id"`
}
