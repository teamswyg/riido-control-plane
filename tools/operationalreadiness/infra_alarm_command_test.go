package main

import (
	"strings"
	"testing"
)

func TestOperationalReadinessClientSurfaceAlarmCommandIsExecutable(t *testing.T) {
	command := readinessCheckByID(t, "otel_xray_client_surface").NextCommand
	for _, token := range []string{
		"terraform -chdir=terraform/riido_ai_server workspace show",
		"terraform -chdir=terraform/riido_ai_server output -raw service_name",
		"gh pr view 116 --repo teamswyg/riido-infra",
		"if ! test -f \"$vars\"",
		"-input=false",
		"-detailed-exitcode",
		"-var-file=\"../../$vars\"",
		"case \"$code\" in 0|2)",
		"aws cloudwatch describe-alarms",
		"go run ./tools/clientalarmpreflight",
		"go run ./tools/clientalarmevidence",
		".preflight.evidence.json",
	} {
		if !strings.Contains(command, token) {
			t.Fatalf("next command missing %q:\n%s", token, command)
		}
	}
}
