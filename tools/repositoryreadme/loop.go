package main

func completeLoop(loop evidenceLoop) bool {
	return loop.Observation != "" &&
		loop.Hypothesis != "" &&
		loop.Execute != "" &&
		loop.Evaluate != "" &&
		loop.Retrospective != ""
}
