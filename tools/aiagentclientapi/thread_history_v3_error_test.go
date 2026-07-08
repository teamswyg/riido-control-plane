package main

import "testing"

func TestVerifyThreadHistoryV3RejectsEndpointAndActionGaps(t *testing.T) {
	t.Parallel()
	assertAIClientAPIError(t, verifyThreadHistoryV3(threadHistoryV3{}), "read endpoint")
	v3 := validThreadHistoryV3()
	v3.SSEEndpoint.Path = "/v3/wrong"
	assertAIClientAPIError(t, verifyThreadHistoryV3(v3), "v2 SSE")
	v3 = validThreadHistoryV3()
	v3.ActionEndpoints = v3.ActionEndpoints[:3]
	assertAIClientAPIError(t, verifyThreadHistoryV3(v3), "action endpoints")
}

func TestVerifyThreadHistoryV3RejectsRuleShapeAndIdentityGaps(t *testing.T) {
	t.Parallel()
	v3 := validThreadHistoryV3()
	v3.ImplementationSteps[0].Detail = ""
	assertAIClientAPIError(t, verifyThreadHistoryV3(v3), "detail is required")
	v3 = validThreadHistoryV3()
	v3.ResponseShapes[1].Fields = []string{"thread_id"}
	assertAIClientAPIError(t, verifyThreadHistoryV3(v3), "missing field")
	v3 = validThreadHistoryV3()
	v3.IdentityRules = []identityRule{{Name: "thread_id"}}
	assertAIClientAPIError(t, verifyThreadHistoryV3(v3), "identity \"conversation_id\"")
}

func TestVerifyThreadHistoryV3RejectsOrderingMutationAndTerminalGaps(t *testing.T) {
	t.Parallel()
	v3 := validThreadHistoryV3()
	v3.OrderingRules = nil
	assertAIClientAPIError(t, verifyThreadHistoryV3(v3), "ordering rule")
	v3 = validThreadHistoryV3()
	v3.MutationRules = nil
	assertAIClientAPIError(t, verifyThreadHistoryV3(v3), "mutation rule")
	v3 = validThreadHistoryV3()
	v3.TerminalStates = []string{"completed"}
	assertAIClientAPIError(t, verifyThreadHistoryV3(v3), "terminal state")
}
