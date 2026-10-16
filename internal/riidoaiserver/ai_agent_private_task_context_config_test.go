package riidoaiserver

import "testing"

func TestPrivateTaskContextGraphQLEndpointValidation(t *testing.T) {
	for _, tc := range []struct {
		name, graphql string
		wantError     bool
	}{
		{"fallback", "", false},
		{"explicit", "https://metadata.invalid/graphql", false},
		{"userinfo", "https://fake-user:fake-secret@metadata.invalid/graphql", true},
		{"query", "https://metadata.invalid/graphql?fake-token=value", true},
		{"fragment", "https://metadata.invalid/graphql#fake-value", true},
		{"invalid_scheme", "file:///graphql", true},
		{"missing_host", "https:///graphql", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client, err := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: "https://rest.invalid", GraphQLURL: tc.graphql})
			if (err != nil) != tc.wantError {
				t.Fatal("GraphQL endpoint validation result changed")
			}
			if err != nil {
				return
			}
			want := tc.graphql
			if want == "" {
				want = "https://rest.invalid"
			}
			if client.baseURL != "https://rest.invalid" || client.graphqlURL != want {
				t.Fatal("REST or GraphQL endpoint changed unexpectedly")
			}
		})
	}
}
