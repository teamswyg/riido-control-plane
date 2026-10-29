package riidoaiserver

import "time"

type AssignmentProjection struct {
	Assignment        Assignment
	LastEventSeq      int64
	Recovery          *recoveryWatermark          `json:"-"`
	EventACKs         []assignmentEventACKReceipt `json:"-"`
	ContextTerminalAt time.Time                   `json:"-"`
}
