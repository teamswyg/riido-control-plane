package riidoaiserver

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPAIAgentClientEventsErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		method string
		token  string
		scopes []string
		store  AIAgentClientStore
		want   int
	}{
		{
			name:   "missing store",
			method: http.MethodGet,
			token:  "user-token",
			want:   http.StatusServiceUnavailable,
		},
		{
			name:   "method not allowed",
			method: http.MethodPost,
			token:  "user-token",
			store:  NewDevelopmentAIAgentClientStore(),
			want:   http.StatusMethodNotAllowed,
		},
		{
			name:   "missing auth",
			method: http.MethodGet,
			store:  NewDevelopmentAIAgentClientStore(),
			want:   http.StatusUnauthorized,
		},
		{
			name:   "forbidden scope",
			method: http.MethodGet,
			token:  "user-token",
			scopes: []string{"component-task:task-a:read"},
			store:  NewDevelopmentAIAgentClientStore(),
			want:   http.StatusForbidden,
		},
		{
			name:   "subscriber error",
			method: http.MethodGet,
			token:  "user-token",
			store: clientEventsSubscriberErrorStore{
				DevelopmentAIAgentClientStore: NewDevelopmentAIAgentClientStore(),
				err:                           errors.New("subscribe failed"),
			},
			want: http.StatusBadRequest,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := newClientEventsErrorTestServer(t, tc.scopes, tc.store)
			req := httptest.NewRequest(tc.method, clientEventsErrorTestPath(), nil)
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)
			if resp.Code != tc.want {
				t.Fatalf("events status=%d want=%d body=%s", resp.Code, tc.want, resp.Body.String())
			}
		})
	}
}
