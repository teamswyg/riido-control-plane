package main

import (
	"context"
	"time"

	srv "github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func buildAssignmentLongPollWait(config) (pressureOperation, error) {
	store := srv.NewStore()
	req := srv.PollRequest{
		DaemonID:  "daemon-pressure",
		DeviceID:  "device-pressure",
		RuntimeID: "runtime-pressure",
		WaitMs:    2,
	}
	return pressureOperation{
		run: func() error {
			_, err := store.WaitForAssignment(context.Background(), "agent-pressure", req, 2*time.Millisecond, time.Millisecond)
			return err
		},
		cleanup: func() {
			store.Close()
		},
	}, nil
}
