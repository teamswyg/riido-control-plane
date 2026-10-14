package main

var checkKindSpecs = []checkKindSpec{
	{
		kind:    "claim",
		key:     idProofKey,
		surface: surfaceFromID(claimProofSurface),
		verify:  verifyIndexed(verifyClaimCheck),
	},
	{
		kind:    "loop",
		key:     idProofKey,
		surface: surfaceFromID(loopProofSurface),
		verify:  verifyIndexed(verifyLoopCheck),
	},
	{
		kind:    "workflow",
		key:     workflowProofKey,
		surface: surfaceFromCheck(workflowProofSurface),
		verify:  verifyRooted(verifyWorkflowCheck),
	},
	{
		kind:    "graph_chain",
		key:     idProofKey,
		surface: surfaceFromID(graphChainProofSurface),
		verify:  verifyIndexed(verifyGraphChainCheck),
	},
	{
		kind:    "graph_edge",
		key:     graphEdgeProofKey,
		surface: surfaceFromCheck(graphEdgeProofSurface),
		verify:  verifyIndexed(verifyGraphEdgeCheck),
	},
	{
		kind:    "graph_summary",
		key:     graphSummaryProofKey,
		surface: surfaceFromIndex(graphSummaryProofSurface),
		verify:  verifyIndexed(verifyGraphSummaryCheck),
	},
	{
		kind:    "harness_summary",
		key:     harnessSummaryProofKey,
		surface: surfaceFromIndex(harnessSummaryProofSurface),
		verify:  verifyIndexed(verifyHarnessSummaryCheck),
	},
	{
		kind:    "pre_commit_hook",
		key:     idProofKey,
		surface: surfaceFromID(preCommitHookProofSurface),
		verify:  verifyIndexed(verifyPreCommitHookCheck),
	},
}
