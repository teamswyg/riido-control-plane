#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

base_ref="${RIIDO_LOOP_REGISTRY_IMPACT_BASE:-origin/main}"
if ! git rev-parse --verify "$base_ref" >/dev/null 2>&1; then
  echo "loopregistry pre-commit: missing base ref '$base_ref'." >&2
  echo "Run 'git fetch origin main' or set RIIDO_LOOP_REGISTRY_IMPACT_BASE." >&2
  exit 1
fi

default_evidence_out="$(git rev-parse --git-path riido-loop-registry-precommit-evidence.json)"
evidence_out="${RIIDO_LOOP_REGISTRY_EVIDENCE_OUT:-$default_evidence_out}"
go run ./tools/loopregistry \
  -check-doc \
  -impact-base "$base_ref" \
  -evidence-out "$evidence_out"
