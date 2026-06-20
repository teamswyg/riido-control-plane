package deploypolicy

import (
	"strings"
	"testing"
)

func requireRedactedSummaryArtifact(t *testing.T, body, artifact string) {
	t.Helper()
	requireContains(t, body, "go run ./tools/liveworkflowevidence")
	requireContains(t, body, "actions/upload-artifact@v4")
	requireContains(t, body, "name: "+artifact)
	requireContains(t, body, "path: out/"+artifact+".json")
	for _, forbidden := range liveHandoffPathNeedles() {
		if strings.Contains(extractArtifactBlock(body, artifact), forbidden) {
			t.Fatalf("redacted artifact block leaks live handoff path %q", forbidden)
		}
	}
}

func liveHandoffPathNeedles() []string {
	return []string{
		"$RUNNER_TEMP",
		"riido-image-uri",
		"riido-task-definition-arn",
		"task-definition.current.json",
		"task-definition.next.json",
		"codedeploy-appspec.json",
		"codedeploy-deployment.json",
	}
}

func extractArtifactBlock(body, artifact string) string {
	index := strings.Index(body, "name: "+artifact)
	if index < 0 {
		return body
	}
	start := strings.LastIndex(body[:index], "- uses: actions/upload-artifact")
	if start < 0 {
		start = index
	}
	end := strings.Index(body[index:], "\n      - ")
	if end < 0 {
		return body[start:]
	}
	return body[start : index+end]
}
