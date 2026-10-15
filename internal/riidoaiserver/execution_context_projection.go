package riidoaiserver

import "github.com/teamswyg/riido-contracts/assignment"

func executionContextFromMetadata(metadata map[string]string) *assignment.ExecutionContext {
	return assignment.ExecutionContextFromMetadata(metadata)
}

func copyExecutionContext(in *assignment.ExecutionContext) *assignment.ExecutionContext {
	if in == nil {
		return nil
	}
	out := in.Sanitized()
	return &out
}

func executionContextsEqual(left, right *assignment.ExecutionContext) bool {
	if left == nil || right == nil {
		return left == right
	}
	return left.Sanitized() == right.Sanitized()
}
