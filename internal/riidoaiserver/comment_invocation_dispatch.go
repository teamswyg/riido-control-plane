package riidoaiserver

import (
	"context"
	"errors"
	"strings"

	"github.com/teamswyg/riido-contracts/assignment"
)

func (s Server) dispatchCommentMentions(ctx context.Context, event TaskEvent) error {
	if s.aiAgent == nil || event.CreatedBy == "" || event.WorkspaceID == "" || event.Type != EventAssignmentCompleted || event.State != AssignmentCompleted || assignmentMetadataNeedsInput(event.Metadata) || assignment.RecoveryContextFromMetadata(event.Metadata) != nil {
		return nil
	}
	principal := AuthorizationResult{PrincipalID: event.CreatedBy, WorkspaceID: event.WorkspaceID}
	source, err := s.commentInvocationSource(ctx, principal, event)
	if err != nil || source.CommentInvocation == nil {
		return err
	}
	recorder, ok := s.aiAgent.(commentInvocationRecorder)
	if !ok {
		return errors.New("comment invocation recorder unavailable")
	}
	notice := func(key, code, message string) error {
		return recorder.RecordCommentInvocationOutcome(ctx, event.TaskID, source.ThreadID, key, code, message)
	}
	body := source.ResultMessage
	if strings.TrimSpace(body) == "" {
		body = source.Message
	}
	ids, err := commentMentionIDs(body)
	if err != nil {
		return notice("policy", "invalid_mentions", "Riido: Agent 호출을 시작하지 못했습니다. 멘션 형식 또는 댓글 크기·대상 수 제한을 확인한 뒤 다시 요청해 주세요.")
	}
	if len(ids) == 0 {
		return nil
	}
	release, err := s.taskAgentExecutionGate.beginPublication(ctx, event.TaskID)
	if err != nil {
		return err
	}
	defer release()
	// Refresh outcomes after waiting for a concurrent HTTP replay to publish.
	source, err = s.commentInvocationSource(ctx, principal, event)
	if err != nil || source.CommentInvocation == nil {
		return err
	}
	call := copyCommentInvocation(source.CommentInvocation)
	origin, visible := taskThreadProjectionMessage(source)
	if !visible || origin.MessageID == "" {
		return errors.New("committed calling message unavailable")
	}
	call.OriginAssignmentID, call.OriginThreadID, call.OriginMessageID = source.AssignmentID, source.ThreadID, origin.MessageID
	if call.Outcomes["policy"] != "" {
		return nil
	}
	if len(commentMentionLimitReasons(call, "", len(ids))) > 0 {
		return notice("policy", "call_limit", "Riido: 자동 Agent 호출 한도에 도달했습니다. 같은 댓글에 새 요청을 남기면 이어갈 수 있습니다.")
	}
	if s.config.CommentDelivery.Key == "" || s.config.CommentDelivery.Validate() != nil {
		return errors.New("comment evidence service unavailable")
	}
	for index, target := range ids {
		if call.Outcomes["target:"+target] != "" {
			continue
		}
		stop, err := s.dispatchCommentTarget(ctx, principal, event, call, body, target, commentChildBudget(call.Remaining, len(ids), index), notice)
		if err != nil || stop {
			return err
		}
	}
	return nil
}
