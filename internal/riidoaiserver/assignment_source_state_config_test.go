package riidoaiserver

import "testing"

func TestAgentSourceStateConfigValidation(t *testing.T) {
	for name, config := range map[string]AgentSourceStateConfig{
		"off":      {},
		"https":    {BaseURL: "https://api.example.com", Key: sourceStateTestKey},
		"loopback": {BaseURL: "http://127.0.0.1:8080", Key: sourceStateTestKey},
	} {
		if err := config.Validate(); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}
	for name, config := range map[string]AgentSourceStateConfig{
		"url only":    {BaseURL: "https://api.example.com"},
		"key only":    {Key: sourceStateTestKey},
		"short key":   {BaseURL: "https://api.example.com", Key: "short"},
		"remote http": {BaseURL: "http://api.example.com", Key: sourceStateTestKey},
		"query":       {BaseURL: "https://api.example.com?x=1", Key: sourceStateTestKey},
		"userinfo":    {BaseURL: "https://user@api.example.com", Key: sourceStateTestKey},
	} {
		if config.Validate() == nil {
			t.Fatalf("%s accepted", name)
		}
	}
}
