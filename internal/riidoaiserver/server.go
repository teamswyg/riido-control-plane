package riidoaiserver

import "time"

const (
	aiAgentTokenHeader = "X-Riido-Ai-Agent-Token"
	deviceIDHeader     = "X-Riido-Device-Id"
	deviceSecretHeader = "X-Riido-Device-Secret"
)

// sseKeepaliveInterval is how often an idle SSE stream emits a comment line to
// keep the connection alive through ALB/proxy idle timeouts.
const sseKeepaliveInterval = 15 * time.Second
