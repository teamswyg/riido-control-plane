package riidoaiserver

type AssignmentProjection struct {
	Assignment   Assignment
	LastEventSeq int64
	Recovery     *recoveryWatermark          `json:"-"`
	EventACKs    []assignmentEventACKReceipt `json:"-"`
}
