package main

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"
)

func TestCLIAgentReadConfigDefaultOffAndSeparateServiceAuthority(t *testing.T) {
	clearRiidoAIServerEnv(t)
	for _, flag := range []string{"", "false"} {
		t.Setenv(envCLIAgentReadEnabled, flag)
		if a, err := cliAgentReadAuthorizerFromEnv(); err != nil || a != nil {
			t.Fatal("default-off changed app authorization")
		}
	}
	t.Setenv(envCLIAgentReadEnabled, "true")
	t.Setenv(envCLIAgentReadMonolithOrigin, "https://monolith.example.test")
	t.Setenv(envCLIAgentReadResource, "https://cp.example.test")
	token := strings.Repeat("t", 32)
	t.Setenv(envCLIAgentReadServiceToken, token)
	if a, err := cliAgentReadAuthorizerFromEnv(); err != nil || a == nil {
		t.Fatal("valid explicit configuration failed", err)
	}
	for _, tc := range []struct{ key, value string }{
		{envCLIAgentReadEnabled, "1"},
		{envCLIAgentReadMonolithOrigin, "http://localhost"},
		{envCLIAgentReadResource, "https://monolith.example.test"},
		{envCLIAgentReadServiceToken, "short"},
		{envExternalAuthzAPIKey, token},
		{envCLIThreadReadServiceToken, token},
		{envCLIThreadReadServiceToken, " " + token + " "},
		{envAuthzTokensJSON, `[{"principal_id":"service:riido-monolith","token":"` + token + `","scopes":["cli-installation:policy-read"]}]`},
		{envAuthzTokensJSON, `[{"principal_id":"service:riido-monolith","token_sha256":"` + fmt.Sprintf("%x", sha256.Sum256([]byte(token))) + `","scopes":["cli-installation:policy-read"]}]`},
		{envAuthzTokensJSON, `[{"principal_id":"service:riido-monolith","token":" ` + token + ` ","scopes":["cli-installation:policy-read"]}]`},
		{envAuthzTokensJSON, `[{"principal_id":"service:riido-monolith","token_sha256":" ` + fmt.Sprintf("%X", sha256.Sum256([]byte(token))) + ` ","scopes":["cli-installation:policy-read"]}]`},
		{envReviewAccountTokenHash, " " + fmt.Sprintf("%X", sha256.Sum256([]byte(token))) + " "},
		{envExternalAuthzAPIKey, " " + token + " "},
	} {
		t.Run(tc.key, func(t *testing.T) {
			t.Setenv(tc.key, tc.value)
			if _, err := cliAgentReadAuthorizerFromEnv(); err == nil {
				t.Fatal("unsafe explicit configuration accepted")
			}
		})
	}
	config, err := runtimeCLIAgentRead(runtimeConfig{})
	if err != nil || config.CLIAgentReadAuthorizer == nil {
		t.Fatal("runtime wiring missing", err)
	}
}
