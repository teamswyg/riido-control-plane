package riidoaiserver

import "time"

type developmentDaemonControlResult struct {
	CommandID  string
	Daemon     DeviceDaemonRecord
	Message    string
	AcceptedAt time.Time
}
