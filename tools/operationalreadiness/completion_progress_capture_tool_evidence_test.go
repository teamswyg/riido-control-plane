package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestCompletionProgressCaptureToolEvidenceIsRedacted(t *testing.T) {
	body, err := os.ReadFile("../../" + completionProgressCaptureToolEvidence)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"eyJ", "Bearer ", "Authorization", "body\":", "result_message"} {
		if strings.Contains(string(body), forbidden) {
			t.Fatalf("capture tool evidence leaked forbidden token %q", forbidden)
		}
	}
	var evidence struct {
		Redacted bool `json:"redacted"`
		Decision struct {
			Status string `json:"status"`
		} `json:"decision"`
		Tool struct {
			Path string `json:"path"`
		} `json:"tool"`
	}
	if err := json.Unmarshal(body, &evidence); err != nil {
		t.Fatal(err)
	}
	if !evidence.Redacted || evidence.Decision.Status != "partial" {
		t.Fatal("capture tool evidence must stay redacted partial")
	}
	if evidence.Tool.Path != "tools/aiagentthreadsnapshot" {
		t.Fatalf("unexpected tool path %q", evidence.Tool.Path)
	}
}
