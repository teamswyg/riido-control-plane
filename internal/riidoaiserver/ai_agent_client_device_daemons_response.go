package riidoaiserver

type DeviceDaemonListResponse struct {
	SchemaVersion string               `json:"schema_version"`
	DeviceID      string               `json:"device_id"`
	Daemons       []DeviceDaemonRecord `json:"daemons"`
}
