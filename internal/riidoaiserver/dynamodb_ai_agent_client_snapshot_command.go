package riidoaiserver

import "context"

type dynamoDBAIAgentClientSnapshotCommand struct {
	ctx      context.Context
	load     bool
	save     *AIAgentClientSnapshot
	close    bool
	loadDone chan dynamoDBAIAgentClientSnapshotLoadResult
	errDone  chan error
}

type dynamoDBAIAgentClientSnapshotLoadResult struct {
	snapshot AIAgentClientSnapshot
	ok       bool
	err      error
}
