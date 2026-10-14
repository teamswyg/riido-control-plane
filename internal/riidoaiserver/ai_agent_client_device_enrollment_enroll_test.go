package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
)

func (f *deviceEnrollmentHTTPFixture) enrollAndRotateDeviceCredential() {
	t := f.t
	enrollReq := httptest.NewRequest(http.MethodPost, "/v2/desktop/workspaces/workspace-alpha/devices/enroll", strings.NewReader(`{"display_name":"JY MacBook","platform":"darwin","app_version":"0.0.0"}`))
	enrollReq.Header.Set(aiAgentTokenHeader, "ai-agent-token")
	enrollResp := httptest.NewRecorder()
	f.server.ServeHTTP(enrollResp, enrollReq)
	if enrollResp.Code != http.StatusCreated {
		t.Fatalf("enroll status=%d body=%s", enrollResp.Code, enrollResp.Body.String())
	}
	if err := json.Unmarshal(enrollResp.Body.Bytes(), &f.enrollment); err != nil {
		t.Fatalf("enroll json: %v", err)
	}
	if f.enrollment.SchemaVersion != DeviceCredentialSchemaVersion ||
		f.enrollment.DeviceID == "" ||
		f.enrollment.DeviceSecret == "" ||
		f.enrollment.OwnerPrincipalID != "user-1" ||
		f.enrollment.WorkspaceID != "workspace-alpha" ||
		f.enrollment.DisplayName != "JY MacBook" {
		t.Fatalf("enrollment = %+v", f.enrollment)
	}
	firstDeviceSecret := f.enrollment.DeviceSecret
	rotateReq := httptest.NewRequest(http.MethodPost, "/v2/desktop/workspaces/workspace-alpha/devices/enroll", strings.NewReader(`{"device_id":"`+f.enrollment.DeviceID+`","display_name":"JY MacBook","platform":"darwin","app_version":"0.0.1"}`))
	rotateReq.Header.Set(aiAgentTokenHeader, "ai-agent-token")
	rotateResp := httptest.NewRecorder()
	f.server.ServeHTTP(rotateResp, rotateReq)
	if rotateResp.Code != http.StatusCreated {
		t.Fatalf("rotate status=%d body=%s", rotateResp.Code, rotateResp.Body.String())
	}
	var rotated EnrollDeviceResponse
	if err := json.Unmarshal(rotateResp.Body.Bytes(), &rotated); err != nil {
		t.Fatalf("rotate json: %v", err)
	}
	if rotated.DeviceID != f.enrollment.DeviceID || rotated.DeviceSecret == "" || rotated.DeviceSecret == firstDeviceSecret ||
		rotated.OwnerPrincipalID != f.enrollment.OwnerPrincipalID || rotated.WorkspaceID != f.enrollment.WorkspaceID {
		t.Fatalf("rotated enrollment = %+v original=%+v", rotated, f.enrollment)
	}
	if _, err := f.aiAgentStore.AuthorizeDeviceCredential(context.Background(), f.enrollment.DeviceID, firstDeviceSecret, AuthorizationRequest{Resource: AuthorizationResourceAgent, Action: AuthorizationActionPoll}); !errors.Is(err, ErrAuthorizationUnauthenticated) {
		t.Fatalf("old rotated device secret must be rejected, got %v", err)
	}
	f.enrollment = rotated
}
