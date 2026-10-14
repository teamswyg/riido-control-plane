package main

const (
	refreshCommandsSchema = "riido-control-plane-loop-refresh-commands.v1"
	dispatchPlanSchema    = "riido-control-plane-loop-refresh-dispatch-plan.v1"
	candidateSchema       = "riido-control-plane-closed-loop-candidates.v1"
)

const (
	dispatchSourceID          = "loop-refresh-dispatch"
	dispatchSourceWorkflow    = ".github/workflows/loop-refresh-dispatch.yml"
	dispatchSummaryArtifact   = "loop-refresh-dispatch-plan"
	dispatchCandidateArtifact = "loop-refresh-dispatch-closed-loop-candidates"
	dispatchHarnessLoop       = "closed_loop_candidate"
	dispatchPromotionTarget   = "closed_loop_candidate"
)
