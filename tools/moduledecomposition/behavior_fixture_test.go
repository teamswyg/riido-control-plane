package main

const moduleDecompositionFixtureManifest = `{
  "schema_version": "fixture.v1",
  "id": "fixture-module-decomposition",
  "title": "Fixture Module Decomposition",
  "riido_task": "fixture",
  "generated_doc": "docs/module-decomposition.md",
  "workflow": ".github/workflows/module-decomposition.yml",
  "evidence_artifact": "module-decomposition-evidence",
  "module_path": "example.com/fixture",
  "source_roots": ["cmd", "internal", "tools"],
  "forbidden_imports": ["net/http"],
  "file_line_budget": {
    "target_lines": 3,
    "sample_limit": 5,
    "hotspot_limit": 5,
    "max_files_over_target": 1,
    "max_file_lines": 5,
    "hotspot_limits": [
      {"path": "cmd/app", "max_files": 1, "max_lines": 5, "max_total_over_target_lines": 2}
    ]
  },
  "packages": [
    {"path": "cmd/app", "kind": "runtime", "role": "entrypoint", "must_not_own": "domain"},
    {"path": "internal/app", "kind": "internal", "role": "domain", "must_not_own": "cli"},
    {"path": "tools/check", "kind": "tool", "role": "verification", "must_not_own": "runtime"}
  ],
  "rules": ["runtime packages must not own domain logic"],
  "loop": {
    "observation": "fixture observation",
    "hypothesis": "fixture hypothesis",
    "execute": "fixture execute",
    "evaluate": "fixture evaluate",
    "retrospective": "fixture retrospective"
  }
}
`
