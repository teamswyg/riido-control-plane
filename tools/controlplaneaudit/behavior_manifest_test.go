package main

const controlPlaneAuditFixtureManifest = `{
  "schema_version": "riido-control-plane-high-traffic-audit.v1",
  "id": "control-plane-high-traffic-audit",
  "title": "Fixture Control Plane High Traffic Audit",
  "generated_doc": "docs/control-plane-high-traffic-audit.md",
  "workflow": ".github/workflows/control-plane-performance.yml",
  "evidence_artifact": "control-plane-high-traffic-audit",
  "race_artifact": "control-plane-race",
  "evidence_tool": "tools/controlplaneaudit",
  "benchmark_command": "go test ./internal/server -run '^$' -bench BenchmarkHotPath -benchmem -count=1",
  "local_pressure_command": "go run ./tools/controlplanepressure -duration 500ms -candidate-out out/candidates.json -evidence-out out/local-pressure.json",
  "manual_pressure_command": "go run ./tools/controlplanepressure -duration 5s -concurrency 1,8,32,128 -candidate-out out/manual-candidates.json -evidence-out out/manual-pressure.json",
  "local_pprof_command": "go run ./tools/controlplanepressure -duration 5s -pprof-dir out/pprof -candidate-out out/pprof-candidates.json -evidence-out out/pprof-pressure.json",
  "race_command": "go test -race ./internal/server -run TestHotPath -count=1",
  "pprof_commands": [
    "go tool pprof -top http://127.0.0.1:6060/debug/pprof/profile?seconds=30",
    "go tool pprof -top http://127.0.0.1:6060/debug/pprof/heap",
    "curl -fsS http://127.0.0.1:6060/debug/pprof/goroutine?debug=1"
  ],
  "required_categories": ["endpoint_hot_path"],
  "surfaces": [{
    "id": "fixture_hot_path",
    "category": "endpoint_hot_path",
    "risk": "fixture risk",
    "files": ["internal/server/hot.go"],
    "patterns": ["sync.Mutex", "Lock()", "make(chan", "Query"],
    "candidate": "fixture candidate"
  }],
  "assertions": ["fixture assertion"],
  "loop": {
    "observation": "fixture observation",
    "hypothesis": "fixture hypothesis",
    "execute": "fixture execute",
    "evaluate": "fixture evaluate",
    "retrospective": "fixture retrospective"
  }
}
`
