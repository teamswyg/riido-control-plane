#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

base_ref="${RIIDO_EVIDENCE_GRAPH_IMPACT_BASE:-origin/main}"
if ! git rev-parse --verify "$base_ref" >/dev/null 2>&1; then
  echo "evidencegraph pre-commit: missing base ref '$base_ref'." >&2
  echo "Run 'git fetch origin main' or set RIIDO_EVIDENCE_GRAPH_IMPACT_BASE." >&2
  exit 1
fi

default_evidence_out="$(git rev-parse --git-path riido-evidence-graph-precommit-evidence.json)"
evidence_out="${RIIDO_EVIDENCE_GRAPH_EVIDENCE_OUT:-$default_evidence_out}"
go run ./tools/evidencegraph \
  -check-doc \
  -impact-base "$base_ref" \
  -evidence-out "$evidence_out"
