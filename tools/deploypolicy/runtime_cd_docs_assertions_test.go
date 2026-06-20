package deploypolicy

import "testing"

func assertRuntimeCDDocs(t *testing.T, p runtimeCDOwnership, f runtimeCDDocFixture) {
	t.Helper()
	assertRuntimeCDDocBody(t, f.Doc)
	assertRuntimeCDBoundaryDoc(t, f.Boundary)
	assertRuntimeCDRepoDocs(t, f.Readme, f.Domain, f.Migration)
	requireNotContains(t, f.Doc+"\n"+f.Boundary+"\n"+f.Integration, "public image digest")
	requireContains(t, p.DependencyDirection.TopDown, "control-plane")
	requireContains(t, p.DependencyDirection.BottomUp, "Infra")
	for _, body := range []string{f.Doc, f.Boundary, f.Integration} {
		requireContains(t, body, "CodeDeploy")
		requireContains(t, body, "riido-control-plane")
		requireContains(t, body, "riido-infra")
	}
}

func assertRuntimeCDDocBody(t *testing.T, doc string) {
	t.Helper()
	for _, want := range runtimeCDDocPhrases() {
		requireContains(t, doc, want)
	}
}

func assertRuntimeCDBoundaryDoc(t *testing.T, boundary string) {
	t.Helper()
	for _, want := range runtimeCDBoundaryPhrases() {
		requireContains(t, boundary, want)
	}
}
