package main

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"
)

func TestCLIThreadReadConfigDefaultOffAndSeparateServiceAuthority(t *testing.T) {
	clearRiidoAIServerEnv(t)
	for _, flag := range []string{"", "false"} {
		t.Setenv(envCLIThreadReadEnabled, flag)
		if a, err := cliThreadReadAuthorizerFromEnv(); err != nil || a != nil {
			t.Fatal("default-off changed app authorization")
		}
	}
	t.Setenv(envCLIThreadReadEnabled, "true")
	t.Setenv(envCLIThreadReadMonolithOrigin, "https://monolith.example.test")
	t.Setenv(envCLIThreadReadResource, "https://cp.example.test")
	token := strings.Repeat("t", 32)
	t.Setenv(envCLIThreadReadServiceToken, token)
	if a, err := cliThreadReadAuthorizerFromEnv(); err != nil || a == nil {
		t.Fatal("valid explicit configuration failed", err)
	}
	for _, tc := range []struct{ key, value string }{
		{envCLIThreadReadEnabled, "1"},
		{envCLIThreadReadMonolithOrigin, "http://localhost"},
		{envCLIThreadReadResource, "https://monolith.example.test"},
		{envCLIThreadReadServiceToken, "short"},
		{envExternalAuthzAPIKey, token},
		{envAuthzTokensJSON, `[{"principal_id":"service:riido-monolith","token":"` + token + `","scopes":["cli-installation:policy-read"]}]`},
		{envAuthzTokensJSON, `[{"principal_id":"service:riido-monolith","token_sha256":"` + fmt.Sprintf("%x", sha256.Sum256([]byte(token))) + `","scopes":["cli-installation:policy-read"]}]`},
		{envAuthzTokensJSON, `[{"principal_id":"service:riido-monolith","token":" ` + token + ` ","scopes":["cli-installation:policy-read"]}]`},
		{envAuthzTokensJSON, `[{"principal_id":"service:riido-monolith","token_sha256":" ` + fmt.Sprintf("%X", sha256.Sum256([]byte(token))) + ` ","scopes":["cli-installation:policy-read"]}]`},
		{envReviewAccountTokenHash, " " + fmt.Sprintf("%X", sha256.Sum256([]byte(token))) + " "},
		{envExternalAuthzAPIKey, " " + token + " "},
	} {
		t.Run(tc.key, func(t *testing.T) {
			t.Setenv(tc.key, tc.value)
			if _, err := cliThreadReadAuthorizerFromEnv(); err == nil {
				t.Fatal("unsafe explicit configuration accepted")
			}
		})
	}
	config, err := runtimeCLIThreadRead(runtimeConfig{})
	if err != nil || config.CLIThreadReadAuthorizer == nil {
		t.Fatal("runtime wiring missing", err)
	}
}
