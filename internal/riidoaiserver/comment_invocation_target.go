package riidoaiserver

import (
	"context"
	"errors"
	"fmt"
	"slices"
)

func (s Server) dispatchCommentTarget(ctx context.Context, principal AuthorizationResult, event TaskEvent, call *CommentInvocation, body, target string, budget int, notice func(string, string, string) error) (bool, error) {
	key := "target:" + target
	if slices.Contains(commentMentionLimitReasons(call, target, 1), "cycle") {
		return false, notice(key, "cycle", "Riido: 자기 자신 또는 앞서 호출한 Agent로 이어지는 반복 호출을 중단했습니다.")
	}
	evidence, err := s.readCommentDeliveryEvidence(ctx, call.EvidenceEventID)
	if errors.Is(err, ErrAuthorizationForbidden) {
		return true, notice("policy", "access_revoked", "Riido: 원댓글 또는 요청자의 현재 접근 권한을 확인할 수 없어 자동 호출을 중단했습니다.")
	}
	if err != nil {
		return true, err
	}
	if !commentInvocationEvidenceMatches(evidence, event.TaskID, call) {
		return true, notice("policy", "scope_mismatch", "Riido: 원래 요청과 대화 연결이 일치하지 않아 자동 호출을 중단했습니다.")
	}
	agents, err := s.aiAgent.ListAIAgentTaskAssignableAgents(ctx, principal, event.TaskID)
	if err != nil {
		return true, err
	}
	selected, err := selectAssignableAgent(agents.Agents, target)
	if err != nil {
		return false, notice(key, "target_unavailable", "Riido: 멘션한 Agent를 현재 권한으로 호출할 수 없습니다. 사용 가능한 Agent를 선택해 다시 요청해 주세요.")
	}
	requestID := commentInvocationRequestID(event.AssignmentID, target)
	// A previous publication may have succeeded before its outcome could persist.
	receipt, found, err := s.findTaskRequestReceipt(ctx, principal, event.TaskID, AssignAIAgentTaskRequest{AgentID: target, RequestID: requestID})
	if err != nil {
		return true, err
	}
	if found {
		if err := s.ensureCommentInvocationContext(ctx, call, receipt.AssignmentID); err != nil {
			return true, err
		}
		return false, notice(key, "accepted", "")
	}
	_, runtime, runtimeErr := s.resolveAgentRuntimeFact(target)
	if runtimeErr != nil || runtime.Availability == RuntimeAvailabilityOffline || selected.WorkStatus == AgentWorkStatusOffline {
		return false, notice(key, "runtime_offline", fmt.Sprintf("Riido: %s의 런타임이 연결되어 있지 않아 호출하지 못했습니다. 연결 후 다시 요청해 주세요.", selected.Name))
	}
	evidence.AgentID = target
	child := copyCommentInvocation(call)
	child.Ancestors = append(child.Ancestors, target)
	child.Depth++
	child.Remaining = budget
	child.Outcomes = nil
	request := AssignAIAgentTaskRequest{AgentID: target, RequestID: requestID, Instruction: fmt.Sprintf("Agent %s requested work in this comment conversation:\n%s", event.AgentID, body), SourceCommentID: evidence.CommentID, SourceMessageID: call.OriginMessageID, VerifiedTaskContext: &evidence.Context, commentInvocation: child, skipSourceMessage: true}
	_, err = s.deliverCommentInvocation(ctx, principal, evidence, request)
	if errors.Is(err, ErrAIAgentTaskThreadConflict) || errors.Is(err, errNativeSessionUnavailable) || errors.Is(err, ErrAuthorizationForbidden) || errors.Is(err, ErrAIAgentNotFound) {
		return false, notice(key, "continuation_unavailable", fmt.Sprintf("Riido: %s의 기존 세션을 이어서 호출할 수 없습니다. 진행 중인 요청과 연결 상태를 확인한 뒤 다시 요청해 주세요.", selected.Name))
	}
	if err != nil {
		return true, err
	}
	return false, notice(key, "accepted", "")
}
