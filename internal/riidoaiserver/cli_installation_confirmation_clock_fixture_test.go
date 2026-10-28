package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"testing"
)

type confirmationClockTransport func(*http.Request) (*http.Response, error)

func (f confirmationClockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func armConfirmationClockAfterLinkRead(t *testing.T, s *DynamoDBAssignmentOperationStore, identity cliInstallationIdentity, arm func()) {
	t.Helper()
	client := *s.httpClient
	base := client.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	key := cliInstallationLinkKey(identity)
	client.Transport = confirmationClockTransport(func(req *http.Request) (*http.Response, error) {
		var body struct{ Key map[string]map[string]string }
		if req.GetBody != nil {
			bodyReader, err := req.GetBody()
			if err == nil {
				_ = json.NewDecoder(bodyReader).Decode(&body)
				_ = bodyReader.Close()
			}
		}
		response, err := base.RoundTrip(req)
		if err == nil && req.Header.Get("X-Amz-Target") == dynamoDBGetItemTarget &&
			body.Key["pk"]["S"] == key["pk"]["S"] && body.Key["sk"]["S"] == key["sk"]["S"] {
			arm()
		}
		return response, err
	})
	s.httpClient = &client
}
