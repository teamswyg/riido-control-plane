package riidoaiserver

func staleDaemonRuntimeSnapshot(previous DeviceDaemonRecord, previousOK bool, next DeviceDaemonRecord) bool {
	if !previousOK {
		return false
	}
	if previous.PID == 0 || next.PID == 0 || previous.PID == next.PID {
		return false
	}
	if previous.StartedAt.IsZero() || next.StartedAt.IsZero() {
		return false
	}
	return next.StartedAt.Before(previous.StartedAt)
}
