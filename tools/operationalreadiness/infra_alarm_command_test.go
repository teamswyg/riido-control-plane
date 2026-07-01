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
		"-input=false",
		"-detailed-exitcode",
		"-var-file=\"../../$vars\"",
		"aws cloudwatch describe-alarms",
		"go run ./tools/clientalarmevidence",
	} {
		if !strings.Contains(command, token) {
			t.Fatalf("next command missing %q:\n%s", token, command)
		}
	}
}
