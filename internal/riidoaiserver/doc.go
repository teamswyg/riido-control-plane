// Package riidoaiserver owns the public Riido SaaS control-plane domain and
// adapter boundary.
//
// The current migration slices contain stdlib-only assignment contracts and
// DTOs, assignment operation journal ports, agent catalog RBAC, agent catalog
// API ports and HTTP adapter, request authorization adapters, agent/runtime
// binding guards, provider status contracts, the in-memory assignment store
// actor, and HTTP/SSE/metrics/health adapters. This package does not own daemon
// provider process execution, AWS deployment wiring, durable store adapters, or
// production secret values.
package riidoaiserver
