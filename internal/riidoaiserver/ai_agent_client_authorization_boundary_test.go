package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestHTTPAIAgentClientAuthenticatedAccessBoundary(t *testing.T) {
	authorizer := accessBoundaryAuthorizer(t)
	for _, mode := range []string{"development", "persistent"} {
		for _, tc := range []struct {
			name, token, agentID, workspaceID string
			readable, mutable                 bool
		}{
			{"owner private", "owner-token", "agent-private-cursor", defaultAIAgentClientWorkspaceID, true, true},
			{"other private", "other-token", "agent-private-cursor", defaultAIAgentClientWorkspaceID, false, false},
			{"other public", "other-token", "agent-public-openclaw", defaultAIAgentClientWorkspaceID, true, false},
			{"admin private", "admin-token", "agent-private-cursor", defaultAIAgentClientWorkspaceID, true, true},
			{"wrong workspace", "owner-token", "agent-private-cursor", "other-workspace", false, false},
		} {
			t.Run(mode+"/"+tc.name, func(t *testing.T) {
				var store AIAgentClientStore = NewDevelopmentAIAgentClientStore()
				if mode == "persistent" {
					var err error
					store, err = OpenPersistentAIAgentClientStore(t.Context(), store.(*DevelopmentAIAgentClientStore), &memoryAIAgentClientSnapshotStore{})
					if err != nil {
						t.Fatal(err)
					}
				}
				assignments := NewStore()
				t.Cleanup(assignments.Close)
				server := NewServer(ServerConfig{AIAgentClient: store, Assignment: assignments, Authorizer: authorizer}).Handler()
				base := "/v2/client/workspaces/" + tc.workspaceID + "/ai-agent"
				bootstrap := accessBoundaryRequest(t, server, http.MethodGet, base+"/bootstrap", tc.token, "", http.StatusOK)
				var list ClientBootstrapResponse
				if err := json.Unmarshal(bootstrap.Body.Bytes(), &list); err != nil {
					t.Fatal(err)
				}
				if found := containsString(aiAgentIDs(list.Agents), tc.agentID); found != tc.readable {
					t.Fatalf("bootstrap target visible=%v want=%v", found, tc.readable)
				}
				readStatus, mutationStatus := http.StatusNotFound, http.StatusNotFound
				if tc.readable {
					readStatus = http.StatusOK
				}
				if tc.mutable {
					mutationStatus = http.StatusOK
				}
				path := base + "/agents/" + tc.agentID
				accessBoundaryRequest(t, server, http.MethodGet, path+"/dashboard", tc.token, "", readStatus)
				accessBoundaryRequest(t, server, http.MethodGet, path+"/editability", tc.token, "", readStatus)
				accessBoundaryRequest(t, server, http.MethodPatch, path, tc.token, `{"name":"changed"}`, mutationStatus)
				accessBoundaryRequest(t, server, http.MethodPatch, path, tc.token, `{"owner_principal_id":"user-1","roles":["admin"]}`, http.StatusBadRequest)
				accessBoundaryRequest(t, server, http.MethodDelete, path, tc.token, "", mutationStatus)
			})
		}
	}
}
