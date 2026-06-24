package riidoaiserver

import "testing"

func TestClientVisibleTaskThreadTextLocalizesProviderLimitMessages(t *testing.T) {
	t.Parallel()
	cases := []string{
		"token usage limit exceeded",
		"Token quota exceeded while researching",
		"insufficient credits for deeper research",
		"credit limit exceeded",
		"토큰 이용 한도 초과",
		"보유하신 크레딧이 부족합니다.",
	}
	for _, input := range cases {
		if got := clientVisibleTaskThreadText(input); got != clientMessageCloudCreditInsufficient {
			t.Fatalf("clientVisibleTaskThreadText(%q) = %q, want %q", input, got, clientMessageCloudCreditInsufficient)
		}
	}
}

func TestFailureDiagnosticsLocalizesProviderLimitMessages(t *testing.T) {
	t.Parallel()
	diagnostics := failureDiagnosticsFromAssignmentEvent(nil, "token usage limit exceeded")
	if diagnostics == nil {
		t.Fatal("expected diagnostics")
	}
	if diagnostics.Message != clientMessageCloudCreditInsufficient {
		t.Fatalf("diagnostics message = %q, want %q", diagnostics.Message, clientMessageCloudCreditInsufficient)
	}
}
