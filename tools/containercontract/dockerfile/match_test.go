package dockerfile

import "testing"

func TestRunMatchers(t *testing.T) {
	runs := []string{
		"RUN --mount=type=cache,target=/go/pkg/mod go mod download",
		"RUN go build -trimpath -ldflags=-s -o /out/app ./cmd/app",
	}
	if !HasModuleDownloadRun(runs, "go mod download") {
		t.Fatal("module download run not found")
	}
	if !HasGoBuildRun(runs, "/out/app", "./cmd/app", true, []string{"-ldflags=-s"}) {
		t.Fatal("go build run not found")
	}
	if !RunHasCacheMount(runs, "go mod download", "/go/pkg/mod") {
		t.Fatal("cache mount not found")
	}
	if HasGoBuildRun(runs, "/out/other", "./cmd/app", true, nil) {
		t.Fatal("unexpected build output matched")
	}
}

func TestHasCopy(t *testing.T) {
	copies := []CopyInstruction{{From: "build", Src: "/out/app", Dst: "/app"}}
	if !HasCopy(copies, "build", "/out/app", "/app") {
		t.Fatal("copy not found")
	}
	if HasCopy(copies, "build", "/out/other", "/app") {
		t.Fatal("unexpected copy matched")
	}
}
