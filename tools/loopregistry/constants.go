package main

const (
	manifestSchema        = "riido-control-plane-loop-registry.v1"
	evidenceSchema        = "riido-control-plane-loop-registry-evidence.v1"
	refreshCommandsSchema = "riido-control-plane-loop-refresh-commands.v1"
	requiredID            = "control-plane-loop-registry"
	defaultManifest       = "docs/30-architecture/loop-registry.riido.json"
)

const (
	kindHarness    = "harness"
	kindClosedLoop = "closed_loop"
)
