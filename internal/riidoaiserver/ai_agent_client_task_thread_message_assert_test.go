package riidoaiserver

import "testing"

func assertTaskThreadMessageRootOnly(
	t *testing.T,
	store AIAgentClientStore,
	root AIAgentTaskActionResponse,
) {
	t.Helper()
	threads, err := store.ListAIAgentTaskThreads(t.Context(), AuthorizationResult{
		PrincipalID: "user-1",
		WorkspaceID: defaultAIAgentClientWorkspaceID,
	}, root.TaskID)
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads: %v", err)
	}
	if len(threads.Threads) != 1 {
		t.Fatalf("threads mutated after rejected message: %+v", threads.Threads)
	}
	thread := threads.Threads[0]
	if thread.ThreadID != root.ThreadID ||
		thread.AssignmentID != root.AssignmentID ||
		thread.AgentID != root.AgentID {
		t.Fatalf("root thread changed: %+v root=%+v", thread, root)
	}
}
