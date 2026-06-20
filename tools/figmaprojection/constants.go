package main

const (
	projectionSchema = "riido-control-plane-figma-ai-agent-projection.v1"
	sourceSchema     = "riido-figma-ai-agent-coverage.v1"
	evidenceSchema   = "riido-control-plane-figma-projection-evidence.v1"
	projectionID     = "figma-ai-agent-control-plane-generated-client-projection"
	sourceID         = "figma-v1-22-ai-agent-ui-coverage"
	requiredTask     = "RIID-4810"
	evidenceTool     = "tools/figmaprojection"
)

const (
	defaultProjection = "docs/30-architecture/figma-ai-agent-control-plane-projection.riido.json"
	defaultSource     = "contracts/ai-agent-client/figma-ai-agent-coverage.riido.json"
)
