package main

var checkKindSpecs = []checkKindSpec{
	{
		kind:    "claim",
		key:     idProofKey,
		surface: claimCheckProofSurface,
		verify:  verifyClaimCheckKind,
	},
	{
		kind:    "loop",
		key:     idProofKey,
		surface: loopCheckProofSurface,
		verify:  verifyLoopCheckKind,
	},
	{
		kind:    "workflow",
		key:     workflowProofKey,
		surface: workflowCheckProofSurface,
		verify:  verifyWorkflowCheckKind,
	},
	{
		kind:    "graph_chain",
		key:     idProofKey,
		surface: graphChainCheckProofSurface,
		verify:  verifyGraphChainCheckKind,
	},
	{
		kind:    "graph_edge",
		key:     graphEdgeProofKey,
		surface: graphEdgeCheckProofSurface,
		verify:  verifyGraphEdgeCheckKind,
	},
	{
		kind:    "graph_summary",
		key:     graphSummaryCheckProofKey,
		surface: graphSummaryCheckProofSurface,
		verify:  verifyGraphSummaryCheckKind,
	},
	{
		kind:    "harness_summary",
		key:     harnessSummaryCheckProofKey,
		surface: harnessSummaryCheckProofSurface,
		verify:  verifyHarnessSummaryCheckKind,
	},
	{
		kind:    "pre_commit_hook",
		key:     idProofKey,
		surface: preCommitHookCheckProofSurface,
		verify:  verifyPreCommitHookCheckKind,
	},
}
