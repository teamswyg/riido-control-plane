package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

type runReadsOwnerAgentStore struct {
	*DevelopmentAIAgentClientStore
}

func (s *runReadsOwnerAgentStore) ListAIAgentTaskAssignableAgents(context.Context, AuthorizationResult, string) (AgentClientListResponse, error) {
	return AgentClientListResponse{Agents: []AgentClientRecord{{AgentID: "agent-a", Name: "A"}, {AgentID: "agent-b", Name: "B"}}}, nil
}

func TestAIAgentRunReadsGroupsChildSourcesByOwner(t *testing.T) {
	server, reader, _ := runReadsTestServer(t, []AssignmentContextRead{
		{SourceKind: "task", SourceID: "task", SourcePart: "body"},
		{SourceKind: "link", SourceID: "task-link", SourcePart: "metadata"},
		{SourceKind: "project", SourceID: "project", SourcePart: "body"},
		{SourceKind: "link", SourceID: "project-link", SourcePart: "metadata"},
		{SourceKind: "comment", SourceID: "comment", SourcePart: "body"},
		{SourceKind: "agent", SourceID: "agent-a", SourcePart: "profile"},
		{SourceKind: "agent", SourceID: "agent-b", SourcePart: "profile"},
	})
	reader.value.Resources = []AIAgentTaskContextResource{{ID: "task-link", Kind: "link", Title: "Task link"}}
	reader.value.Hierarchy.Project = AIAgentTaskContextReference{ID: "project", Title: "Project", Document: &AIAgentTaskContextDocument{Content: "project body"}, Resources: []AIAgentTaskContextResource{{ID: "project-link", Kind: "link", Title: "Project link"}}}
	reader.value.Extended = &AIAgentTaskContextExtended{Comments: []AIAgentTaskContextComment{{ID: "comment", Body: "comment body"}}, CommentStatus: "available", RelatedStatus: "available", MeetingStatus: "available"}
	server.aiAgent = &runReadsOwnerAgentStore{DevelopmentAIAgentClientStore: server.aiAgent.(*DevelopmentAIAgentClientStore)}

	response := runReadsRequest(server, "/v1/client/ai-agent/runs/run/reads")
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	var body AIAgentRunReadsResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	owners := map[string]string{}
	for _, read := range body.Reads {
		if !read.Listed {
			t.Fatalf("cli read must be listed: %+v", read)
		}
		owners[read.Kind+":"+read.ID] = read.OwnerKind + "/" + read.OwnerID
	}
	want := map[string]string{
		"task:task": "/", "link:task-link": "task/task", "project:project": "/", "link:project-link": "project/project",
		"comment:comment": "task/task", "agent:agent-a": "agent_directory/", "agent:agent-b": "agent_directory/",
	}
	for key, owner := range want {
		if owners[key] != owner {
			t.Fatalf("%s owner=%q want %q (all=%v)", key, owners[key], owner, owners)
		}
	}
	if body.SourceTotal != 3 {
		t.Fatalf("source_total=%d want task, project and one Agent list", body.SourceTotal)
	}
}
