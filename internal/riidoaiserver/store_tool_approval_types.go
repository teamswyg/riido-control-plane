package riidoaiserver

type createToolApprovalCmd struct {
	agentID string
	req     ToolApprovalRequest
	reply   chan toolApprovalCreateResult
}

type toolApprovalCreateResult struct {
	approval ToolApprovalRequest
	err      error
}

type listToolApprovalsCmd struct {
	taskID string
	reply  chan toolApprovalListResult
}

type toolApprovalListResult struct {
	approvals []ToolApprovalRequest
	err       error
}

type decideToolApprovalCmd struct {
	taskID   string
	decision ToolApprovalDecision
	reply    chan toolApprovalDecisionResult
}

type readToolApprovalCmd struct {
	agentID      string
	assignmentID string
	approvalID   string
	reply        chan toolApprovalDecisionResult
}

type toolApprovalDecisionResult struct {
	result   ToolApprovalResult
	decision *ToolApprovalDecision
	mutated  bool
	err      error
}

type registerToolApprovalWaiterCmd struct {
	key   string
	reply chan registerToolApprovalWaiterResult
}

type registerToolApprovalWaiterResult struct {
	ch chan struct{}
	id int64
}

type unregisterToolApprovalWaiterCmd struct {
	key   string
	id    int64
	reply chan struct{}
}
