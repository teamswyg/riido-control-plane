package main

func verifyClaimCheckKind(_ string, c check, idx indexes) error {
	return verifyClaimCheck(c, idx)
}

func verifyLoopCheckKind(_ string, c check, idx indexes) error {
	return verifyLoopCheck(c, idx)
}

func verifyWorkflowCheckKind(root string, c check, _ indexes) error {
	return verifyWorkflowCheck(root, c)
}

func verifyGraphChainCheckKind(_ string, c check, idx indexes) error {
	return verifyGraphChainCheck(c, idx)
}

func verifyGraphEdgeCheckKind(_ string, c check, idx indexes) error {
	return verifyGraphEdgeCheck(c, idx)
}

func verifyGraphSummaryCheckKind(_ string, c check, idx indexes) error {
	return verifyGraphSummaryCheck(c, idx)
}

func verifyHarnessSummaryCheckKind(_ string, c check, idx indexes) error {
	return verifyHarnessSummaryCheck(c, idx)
}

func verifyPreCommitHookCheckKind(_ string, c check, idx indexes) error {
	return verifyPreCommitHookCheck(c, idx)
}
