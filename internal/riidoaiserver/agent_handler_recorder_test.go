package riidoaiserver

import "context"

type handlerRecorderStore struct {
	AIAgentClientStore
	progressResp  AgentThreadProgressBatchResponse
	progressErr   error
	progressCalls int
	progressReq   AgentThreadProgressBatchRequest
	eventCalls    int
	eventReq      AgentEventRequest
	event         TaskEvent
}

func (s *handlerRecorderStore) RecordAIAgentThreadProgress(
	_ context.Context,
	_ string,
	req AgentThreadProgressBatchRequest,
) (AgentThreadProgressBatchResponse, error) {
	s.progressCalls++
	s.progressReq = req
	return s.progressResp, s.progressErr
}

func (s *handlerRecorderStore) RecordAIAgentAssignmentEvent(
	_ context.Context,
	_ string,
	req AgentEventRequest,
	event TaskEvent,
) error {
	s.eventCalls++
	s.eventReq = req
	s.event = event
	return nil
}
