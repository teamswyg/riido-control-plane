package riidoaiserver

import (
	"strings"
	"testing"
)

func TestAWSJSONOperationName(t *testing.T) {
	tests := []struct {
		target string
		want   string
	}{
		{target: "DynamoDB_20120810.PutItem", want: "PutItem"},
		{target: "NoDotTarget", want: "NoDotTarget"},
		{target: "", want: "unknown"},
		{target: "Service.", want: "Service."},
	}
	for _, tt := range tests {
		if got := awsJSONOperationName(tt.target); got != tt.want {
			t.Fatalf("awsJSONOperationName(%q) = %q, want %q", tt.target, got, tt.want)
		}
	}
}

func TestAWSJSONAPIErrorDefaultService(t *testing.T) {
	err := awsJSONAPIError{statusCode: 503, body: []byte(" unavailable \n")}
	got := err.Error()
	if !strings.Contains(got, "aws-json api error") ||
		!strings.Contains(got, "status=503") ||
		!strings.Contains(got, `body="unavailable"`) {
		t.Fatalf("error = %q", got)
	}
}
