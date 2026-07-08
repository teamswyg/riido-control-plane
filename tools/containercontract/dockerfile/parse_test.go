package dockerfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseDockerfile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "Dockerfile")
	body := `
ARG GO_VERSION=1.26.2
FROM golang:${GO_VERSION} AS build
WORKDIR /src
ENV CGO_ENABLED=0
RUN --mount=type=cache,target=/go/pkg/mod go mod download
RUN go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/app
FROM scratch
COPY --from=build /out/app /app
EXPOSE 8080/tcp 9090
USER 10001
ENTRYPOINT ["/app"]
`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	parsed, err := Parse(path)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Args["GO_VERSION"] != "1.26.2" {
		t.Fatalf("args = %#v", parsed.Args)
	}
	build := parsed.StageByAlias("build")
	if build == nil || build.Workdir != "/src" || build.Env["CGO_ENABLED"] != "0" {
		t.Fatalf("build stage = %#v", build)
	}
	final := parsed.FinalStage()
	if final == nil || final.User != "10001" || !IntSetEqual(final.Exposes, []int{9090, 8080}) {
		t.Fatalf("final stage = %#v", final)
	}
	if len(final.Entrypoint) != 1 || final.Entrypoint[0] != "/app" {
		t.Fatalf("entrypoint = %#v", final.Entrypoint)
	}
}
