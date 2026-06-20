package deploypolicy

import (
	"strings"
	"testing"
)

func assertPublicRedactionDocs(t *testing.T, f redactionFixture) {
	t.Helper()
	for _, assertion := range publicRedactionDocAssertions(f) {
		requireContains(t, assertion.body, assertion.want)
	}
}

func assertNoPinnedLiveHost(t *testing.T, f redactionFixture) {
	t.Helper()
	for path, body := range publicRedactionBodies(f) {
		if strings.Contains(body, "ai-api.riido.io") {
			t.Fatalf("%s must not pin the live testnet host", path)
		}
	}
}

type docAssertion struct {
	body string
	want string
}
